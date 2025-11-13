package sshclient

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewConnectionPool(t *testing.T) {
	pool := NewConnectionPool()

	assert.NotNil(t, pool)
	assert.NotNil(t, pool.connections)
	assert.Equal(t, 5*time.Minute, pool.maxIdle)
	assert.Equal(t, 30*time.Second, pool.healthCheck)
	assert.Equal(t, 3, pool.maxRetries)
	assert.Equal(t, 1*time.Second, pool.retryDelay)
	assert.Empty(t, pool.connections)
}

func TestMakeKey(t *testing.T) {
	pool := NewConnectionPool()

	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "Standard configuration",
			config: &Config{
				Host: "192.168.1.100",
				Port: "22",
				User: "root",
			},
			expected: "root@192.168.1.100:22",
		},
		{
			name: "Custom port",
			config: &Config{
				Host: "example.com",
				Port: "2222",
				User: "admin",
			},
			expected: "admin@example.com:2222",
		},
		{
			name: "Different username",
			config: &Config{
				Host: "localhost",
				Port: "22",
				User: "testuser",
			},
			expected: "testuser@localhost:22",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := pool.makeKey(tt.config)
			assert.Equal(t, tt.expected, key)
		})
	}
}

func TestIsConnectionAlive_NilClient(t *testing.T) {
	pool := NewConnectionPool()
	alive := pool.isConnectionAlive(nil)
	assert.False(t, alive)
}

func TestReleaseConnection(t *testing.T) {
	pool := NewConnectionPool()
	config := &Config{
		Host: "test-host",
		Port: "22",
		User: "testuser",
	}

	// Create a mock connection
	key := pool.makeKey(config)
	pooledConn := &PooledConnection{
		client:   nil, // Use nil to avoid Close issues
		config:   config,
		lastUsed: time.Now().Add(-1 * time.Minute),
		inUse:    true,
	}

	pool.mu.Lock()
	pool.connections[key] = pooledConn
	pool.mu.Unlock()

	// Release the connection
	pool.ReleaseConnection(config)

	// Verify connection state
	pool.mu.RLock()
	conn := pool.connections[key]
	pool.mu.RUnlock()

	assert.NotNil(t, conn)
	conn.mu.Lock()
	assert.False(t, conn.inUse)
	assert.WithinDuration(t, time.Now(), conn.lastUsed, 2*time.Second)
	conn.mu.Unlock()
}

func TestReleaseConnection_NonExistent(t *testing.T) {
	pool := NewConnectionPool()
	config := &Config{
		Host: "nonexistent",
		Port: "22",
		User: "testuser",
	}

	// Releasing a non-existent connection should not panic
	assert.NotPanics(t, func() {
		pool.ReleaseConnection(config)
	})
}

func TestClose(t *testing.T) {
	pool := NewConnectionPool()

	// Add some mock connections (with different keys)
	for i := 0; i < 3; i++ {
		config := &Config{
			Host: fmt.Sprintf("test-host-%d", i),
			Port: "22",
			User: "testuser",
		}
		key := pool.makeKey(config)
		pooledConn := &PooledConnection{
			client:   nil, // Use nil to avoid panic on Close
			config:   config,
			lastUsed: time.Now(),
			inUse:    false,
		}
		pool.connections[key] = pooledConn
	}

	assert.Equal(t, 3, len(pool.connections))

	// Close the connection pool (client is nil, so Close won't be called)
	pool.Close()

	// Verify all connections have been cleared
	assert.Empty(t, pool.connections)
}

func TestStats(t *testing.T) {
	pool := NewConnectionPool()

	// Add some mock connections
	activeConfig := &Config{Host: "active-host", Port: "22", User: "active"}
	idleConfig := &Config{Host: "idle-host", Port: "22", User: "idle"}

	activeKey := pool.makeKey(activeConfig)
	idleKey := pool.makeKey(idleConfig)

	pool.connections[activeKey] = &PooledConnection{
		client:   nil,
		config:   activeConfig,
		lastUsed: time.Now(),
		inUse:    true,
	}

	pool.connections[idleKey] = &PooledConnection{
		client:   nil,
		config:   idleConfig,
		lastUsed: time.Now(),
		inUse:    false,
	}

	// Get statistics
	stats := pool.Stats()

	assert.Equal(t, 2, stats["total_connections"])
	assert.Equal(t, 1, stats["active_connections"])
	assert.Equal(t, 1, stats["idle_connections"])
	assert.Equal(t, "5m0s", stats["max_idle_duration"])
	assert.Equal(t, "30s", stats["health_check_interval"])
}

func TestStats_EmptyPool(t *testing.T) {
	pool := NewConnectionPool()

	stats := pool.Stats()

	assert.Equal(t, 0, stats["total_connections"])
	assert.Equal(t, 0, stats["active_connections"])
	assert.Equal(t, 0, stats["idle_connections"])
}

func TestCleanup_IdleConnections(t *testing.T) {
	pool := NewConnectionPool()
	pool.maxIdle = 100 * time.Millisecond // Set a very short idle time for testing

	config := &Config{Host: "test-host", Port: "22", User: "testuser"}
	key := pool.makeKey(config)

	// Add an expired idle connection (client set to nil to avoid panic during cleanup)
	pool.connections[key] = &PooledConnection{
		client:   nil,
		config:   config,
		lastUsed: time.Now().Add(-200 * time.Millisecond), // Already expired
		inUse:    false,
	}

	assert.Equal(t, 1, len(pool.connections))

	// Perform cleanup
	pool.cleanup()

	// Verify expired connection has been removed
	assert.Empty(t, pool.connections)
}

func TestCleanup_ActiveConnections(t *testing.T) {
	pool := NewConnectionPool()
	pool.maxIdle = 100 * time.Millisecond

	config := &Config{Host: "test-host", Port: "22", User: "testuser"}
	key := pool.makeKey(config)

	// Add an active connection (should not be cleaned even if expired, unless connection is dead)
	pool.connections[key] = &PooledConnection{
		client:   nil,
		config:   config,
		lastUsed: time.Now().Add(-200 * time.Millisecond),
		inUse:    true, // In use
	}

	assert.Equal(t, 1, len(pool.connections))

	// Perform cleanup
	pool.cleanup()

	// Since isConnectionAlive(nil) returns false, connection will be cleaned
	assert.Empty(t, pool.connections)
}

func TestPooledConnection_ConcurrentAccess(t *testing.T) {
	pool := NewConnectionPool()
	config := &Config{Host: "test-host", Port: "22", User: "testuser"}
	key := pool.makeKey(config)

	pooledConn := &PooledConnection{
		client:   nil,
		config:   config,
		lastUsed: time.Now(),
		inUse:    false,
	}

	pool.mu.Lock()
	pool.connections[key] = pooledConn
	pool.mu.Unlock()

	// Concurrent access test
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			pool.ReleaseConnection(config)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify no panic occurred
	pool.mu.RLock()
	conn := pool.connections[key]
	pool.mu.RUnlock()

	assert.NotNil(t, conn)
	assert.False(t, conn.inUse)
}
