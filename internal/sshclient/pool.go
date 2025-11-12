package sshclient

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// ConnectionPool SSH连接池
type ConnectionPool struct {
	mu          sync.RWMutex
	connections map[string]*PooledConnection
	maxIdle     time.Duration // 最大空闲时间
	healthCheck time.Duration // 健康检查间隔
	maxRetries  int           // 最大重试次数
	retryDelay  time.Duration // 重试延迟
}

// PooledConnection 池化的连接
type PooledConnection struct {
	client     *ssh.Client
	config     *Config
	lastUsed   time.Time
	mu         sync.Mutex
	inUse      bool
	retryCount int
}

var (
	globalPool     *ConnectionPool
	globalPoolOnce sync.Once
)

// GetConnectionPool 获取全局连接池（单例）
func GetConnectionPool() *ConnectionPool {
	globalPoolOnce.Do(func() {
		globalPool = NewConnectionPool()
		// 启动后台健康检查和清理
		go globalPool.startMaintenance()
	})
	return globalPool
}

// NewConnectionPool 创建连接池
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connections: make(map[string]*PooledConnection),
		maxIdle:     5 * time.Minute,  // 5分钟无使用自动关闭
		healthCheck: 30 * time.Second, // 30秒健康检查
		maxRetries:  3,                // 最大重试3次
		retryDelay:  1 * time.Second,  // 重试延迟1秒
	}
}

// GetConnection 从连接池获取或创建连接
func (p *ConnectionPool) GetConnection(config *Config) (*ssh.Client, error) {
	key := p.makeKey(config)

	p.mu.Lock()
	pooledConn, exists := p.connections[key]

	if exists {
		// 检查连接是否还有效
		if p.isConnectionAlive(pooledConn.client) {
			pooledConn.mu.Lock()
			pooledConn.lastUsed = time.Now()
			pooledConn.inUse = true
			pooledConn.retryCount = 0 // 重置重试计数
			pooledConn.mu.Unlock()
			p.mu.Unlock()
			return pooledConn.client, nil
		}

		// 连接失效，移除并重新创建
		pooledConn.client.Close()
		delete(p.connections, key)
	}
	p.mu.Unlock()

	// 创建新连接（带重试机制）
	client, err := p.createConnectionWithRetry(config)
	if err != nil {
		return nil, err
	}

	// 添加到连接池
	pooledConn = &PooledConnection{
		client:     client,
		config:     config,
		lastUsed:   time.Now(),
		inUse:      true,
		retryCount: 0,
	}

	p.mu.Lock()
	p.connections[key] = pooledConn
	p.mu.Unlock()

	return client, nil
}

// ReleaseConnection 释放连接回连接池
func (p *ConnectionPool) ReleaseConnection(config *Config) {
	key := p.makeKey(config)

	p.mu.RLock()
	pooledConn, exists := p.connections[key]
	p.mu.RUnlock()

	if exists {
		pooledConn.mu.Lock()
		pooledConn.inUse = false
		pooledConn.lastUsed = time.Now()
		pooledConn.mu.Unlock()
	}
}

// createConnectionWithRetry 创建连接（带重试）
func (p *ConnectionPool) createConnectionWithRetry(config *Config) (*ssh.Client, error) {
	var lastErr error

	for i := 0; i < p.maxRetries; i++ {
		if i > 0 {
			time.Sleep(p.retryDelay * time.Duration(i)) // 指数退避
		}

		client, err := p.createConnection(config)
		if err == nil {
			return client, nil
		}

		lastErr = err
	}

	return nil, fmt.Errorf("failed after %d retries: %w", p.maxRetries, lastErr)
}

// createConnection 创建单个SSH连接（直接连接，不使用连接池）
func (p *ConnectionPool) createConnection(config *Config) (*ssh.Client, error) {
	sshClient, err := NewSSHClient(config)
	if err != nil {
		return nil, err
	}

	// 使用 connectDirect() 避免递归调用连接池
	if err := sshClient.connectDirect(); err != nil {
		return nil, err
	}

	return sshClient.client, nil
}

// isConnectionAlive 检查连接是否存活
func (p *ConnectionPool) isConnectionAlive(client *ssh.Client) bool {
	if client == nil {
		return false
	}

	// 尝试创建一个session来测试连接
	session, err := client.NewSession()
	if err != nil {
		return false
	}
	session.Close()

	return true
}

// makeKey 生成连接池键
func (p *ConnectionPool) makeKey(config *Config) string {
	return fmt.Sprintf("%s@%s:%s", config.User, config.Host, config.Port)
}

// startMaintenance 启动后台维护任务
func (p *ConnectionPool) startMaintenance() {
	ticker := time.NewTicker(p.healthCheck)
	defer ticker.Stop()

	for range ticker.C {
		p.cleanup()
	}
}

// cleanup 清理过期和失效的连接
func (p *ConnectionPool) cleanup() {
	now := time.Now()
	var toRemove []string

	p.mu.RLock()
	for key, pooledConn := range p.connections {
		pooledConn.mu.Lock()

		// 检查是否超过最大空闲时间且未在使用
		if !pooledConn.inUse && now.Sub(pooledConn.lastUsed) > p.maxIdle {
			toRemove = append(toRemove, key)
		} else if !p.isConnectionAlive(pooledConn.client) {
			// 连接失效
			toRemove = append(toRemove, key)
		}

		pooledConn.mu.Unlock()
	}
	p.mu.RUnlock()

	// 移除失效连接
	if len(toRemove) > 0 {
		p.mu.Lock()
		for _, key := range toRemove {
			if pooledConn, exists := p.connections[key]; exists {
				pooledConn.client.Close()
				delete(p.connections, key)
			}
		}
		p.mu.Unlock()
	}
}

// Close 关闭连接池中的所有连接
func (p *ConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, pooledConn := range p.connections {
		pooledConn.client.Close()
	}

	p.connections = make(map[string]*PooledConnection)
}

// Stats 获取连接池统计信息
func (p *ConnectionPool) Stats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	totalConns := len(p.connections)
	activeConns := 0
	idleConns := 0

	for _, pooledConn := range p.connections {
		pooledConn.mu.Lock()
		if pooledConn.inUse {
			activeConns++
		} else {
			idleConns++
		}
		pooledConn.mu.Unlock()
	}

	return map[string]interface{}{
		"total_connections":     totalConns,
		"active_connections":    activeConns,
		"idle_connections":      idleConns,
		"max_idle_duration":     p.maxIdle.String(),
		"health_check_interval": p.healthCheck.String(),
	}
}
