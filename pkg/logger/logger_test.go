package logger

import (
	"os"
	"testing"
)

func TestLogger(t *testing.T) {
	// 创建测试日志器
	logger := NewLogger(LogLevelDebug, "[TEST] ")

	// 测试不同级别的日志
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warning("This is a warning message")
	logger.Error("This is an error message")
	logger.Success("This is a success message")
	logger.Tip("This is a tip message")
}

func TestFileLogging(t *testing.T) {
	// 创建临时日志文件
	tmpDir := t.TempDir()
	logPath := tmpDir + "/test.log"

	logger := NewLogger(LogLevelInfo, "")
	if err := logger.EnableFileLogging(logPath); err != nil {
		t.Fatalf("Failed to enable file logging: %v", err)
	}
	defer func() { _ = logger.Close() }() //nolint:errcheck // test cleanup

	// 写入日志
	logger.Info("Test log message")
	logger.Success("Test success message")

	// 验证文件存在
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("Log file was not created: %s", logPath)
	}
}

func TestLogRotation(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := tmpDir + "/rotate.log"

	logger := NewLogger(LogLevelInfo, "")
	logger.SetMaxSize(100) // 设置很小的文件大小以触发轮换
	logger.SetMaxFiles(3)

	if err := logger.EnableFileLogging(logPath); err != nil {
		t.Fatalf("Failed to enable file logging: %v", err)
	}
	defer func() { _ = logger.Close() }() //nolint:errcheck // test cleanup

	// 写入足够多的日志以触发轮换
	for i := 0; i < 50; i++ {
		logger.Info("This is a long log message to trigger rotation %d", i)
	}

	// 手动触发轮换
	if err := logger.Rotate(); err != nil {
		t.Errorf("Failed to rotate log: %v", err)
	}

	// 验证备份文件存在
	if _, err := os.Stat(logPath + ".1"); os.IsNotExist(err) {
		t.Errorf("Rotated log file was not created: %s.1", logPath)
	}
}

func TestLogLevels(t *testing.T) {
	logger := NewLogger(LogLevelWarning, "")

	// 设置级别为 Warning，Info 和 Debug 不应输出
	oldLevel := logger.GetLevel()
	if oldLevel != LogLevelWarning {
		t.Errorf("Expected log level %v, got %v", LogLevelWarning, oldLevel)
	}

	// 更改级别
	logger.SetLevel(LogLevelDebug)
	if logger.GetLevel() != LogLevelDebug {
		t.Errorf("Failed to set log level")
	}
}

func TestLogLevelFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", LogLevelDebug},
		{"DEBUG", LogLevelDebug},
		{"info", LogLevelInfo},
		{"INFO", LogLevelInfo},
		{"warning", LogLevelWarning},
		{"warn", LogLevelWarning},
		{"error", LogLevelError},
		{"unknown", LogLevelInfo}, // 默认值
	}

	for _, tt := range tests {
		result := LogLevelFromString(tt.input)
		if result != tt.expected {
			t.Errorf("LogLevelFromString(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}
