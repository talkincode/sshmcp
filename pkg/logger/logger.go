package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// LogLevel 定义日志级别
type LogLevel int

const (
	// LogLevelDebug 调试级别
	LogLevelDebug LogLevel = iota
	// LogLevelInfo 信息级别
	LogLevelInfo
	// LogLevelWarning 警告级别
	LogLevelWarning
	// LogLevelError 错误级别
	LogLevelError
)

const (
	// DefaultLogDir 默认日志目录
	DefaultLogDir = "~/.sshmcp"
	// DefaultLogFile 默认日志文件名
	DefaultLogFile = "sshx.log"
	// DefaultMaxFiles 默认保留的日志文件数量
	DefaultMaxFiles = 3
	// DefaultMaxSize 默认单个日志文件最大大小（10MB）
	DefaultMaxSize = 10 * 1024 * 1024
)

// Logger 统一的日志记录器
type Logger struct {
	mu          sync.RWMutex
	level       LogLevel
	consoleOut  io.Writer // 控制台输出（stderr）
	fileOut     io.Writer // 文件输出
	logFile     *os.File  // 当前日志文件
	logPath     string    // 日志文件路径
	maxSize     int64     // 最大文件大小
	maxFiles    int       // 最大文件数量
	currentSize int64     // 当前文件大小
	prefix      string
	debugLog    *log.Logger
	infoLog     *log.Logger
	warnLog     *log.Logger
	errorLog    *log.Logger
}

var (
	globalLogger     *Logger
	globalLoggerOnce sync.Once
)

// GetLogger 获取全局日志记录器
func GetLogger() *Logger {
	globalLoggerOnce.Do(func() {
		globalLogger = NewLogger(LogLevelInfo, "")
		// 尝试启用文件日志
		if err := globalLogger.EnableFileLogging(""); err != nil {
			// 如果启用文件日志失败，只输出到 stderr
			fmt.Fprintf(os.Stderr, "Warning: Failed to enable file logging: %v\n", err)
		}
	})
	return globalLogger
}

// SetGlobalLogger 设置全局日志记录器
func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}

// NewLogger 创建新的日志记录器
// level: 日志级别
// prefix: 日志前缀
func NewLogger(level LogLevel, prefix string) *Logger {
	logger := &Logger{
		level:      level,
		consoleOut: os.Stderr, // 默认输出到 stderr，不影响 stdout
		prefix:     prefix,
		maxSize:    DefaultMaxSize,
		maxFiles:   DefaultMaxFiles,
	}

	logger.initLoggers()
	return logger
}

// initLoggers 初始化日志输出器
func (l *Logger) initLoggers() {
	// 如果有文件输出，使用 MultiWriter 同时写入控制台和文件
	var output io.Writer
	if l.fileOut != nil {
		output = io.MultiWriter(l.consoleOut, l.fileOut)
	} else {
		output = l.consoleOut
	}

	l.debugLog = log.New(output, l.prefix+"[DEBUG] ", log.LstdFlags)
	l.infoLog = log.New(output, l.prefix+"", log.LstdFlags)
	l.warnLog = log.New(output, l.prefix+"⚠️  ", log.LstdFlags)
	l.errorLog = log.New(output, l.prefix+"❌ ", log.LstdFlags)
}

// EnableFileLogging 启用文件日志
// logPath: 日志文件路径，如果为空则使用默认路径 ~/.sshmcp/sshx.log
func (l *Logger) EnableFileLogging(logPath string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 如果没有指定路径，使用默认路径
	if logPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		logDir := filepath.Join(home, ".sshmcp")
		logPath = filepath.Join(logDir, DefaultLogFile)
	}

	// 展开 ~ 符号
	if logPath[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		logPath = filepath.Join(home, logPath[1:])
	}

	// 创建日志目录
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0750); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// 打开日志文件
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600) //nolint:gosec // controlled log path
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// 获取当前文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		_ = file.Close() //nolint:errcheck // cleanup on error path
		return fmt.Errorf("failed to stat log file: %w", err)
	}

	// 关闭旧的日志文件
	if l.logFile != nil {
		_ = l.logFile.Close() //nolint:errcheck // closing old file
	}

	l.logFile = file
	l.fileOut = file
	l.logPath = logPath
	l.currentSize = fileInfo.Size()

	// 重新初始化 loggers
	l.initLoggers()

	return nil
}

// DisableFileLogging 禁用文件日志
func (l *Logger) DisableFileLogging() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logFile != nil {
		if err := l.logFile.Close(); err != nil {
			return err
		}
		l.logFile = nil
		l.fileOut = nil
		l.logPath = ""
		l.currentSize = 0
		l.initLoggers()
	}

	return nil
}

// rotateNoLock 轮换日志文件（不加锁版本，需要调用者持有锁）
func (l *Logger) rotateNoLock() error {
	if l.logFile == nil || l.logPath == "" {
		return nil
	}

	// 关闭当前文件
	if err := l.logFile.Close(); err != nil {
		return err
	}

	// 轮换文件：sshx.log.2 -> 删除, sshx.log.1 -> sshx.log.2, sshx.log -> sshx.log.1
	for i := l.maxFiles - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", l.logPath, i)
		newPath := fmt.Sprintf("%s.%d", l.logPath, i+1)

		if i == l.maxFiles-1 {
			// 删除最老的文件
			_ = os.Remove(oldPath) //nolint:errcheck // cleanup old files
		} else {
			// 重命名文件
			if _, err := os.Stat(oldPath); err == nil {
				_ = os.Rename(oldPath, newPath) //nolint:errcheck // best effort file rotation
			}
		}
	}

	// 重命名当前日志文件
	firstBackup := fmt.Sprintf("%s.1", l.logPath)
	if err := os.Rename(l.logPath, firstBackup); err != nil {
		// 如果重命名失败，尝试直接创建新文件
		_ = os.Remove(l.logPath) //nolint:errcheck // cleanup on rename failure
	}

	// 创建新的日志文件
	file, err := os.OpenFile(l.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to create new log file: %w", err)
	}

	l.logFile = file
	l.fileOut = file
	l.currentSize = 0
	l.initLoggers()

	return nil
}

// Rotate 手动轮换日志文件
func (l *Logger) Rotate() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.rotateNoLock()
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel 获取当前日志级别
func (l *Logger) GetLevel() LogLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// SetMaxSize 设置最大文件大小
func (l *Logger) SetMaxSize(size int64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.maxSize = size
}

// SetMaxFiles 设置最大文件数量
func (l *Logger) SetMaxFiles(count int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.maxFiles = count
}

// Debug 记录调试信息
func (l *Logger) Debug(format string, args ...interface{}) {
	l.mu.RLock()
	level := l.level
	l.mu.RUnlock()

	if level <= LogLevelDebug {
		l.debugLog.Printf(format, args...)
		l.checkRotation()
	}
}

// Info 记录普通信息
func (l *Logger) Info(format string, args ...interface{}) {
	l.mu.RLock()
	level := l.level
	l.mu.RUnlock()

	if level <= LogLevelInfo {
		l.infoLog.Printf(format, args...)
		l.checkRotation()
	}
}

// Warning 记录警告信息
func (l *Logger) Warning(format string, args ...interface{}) {
	l.mu.RLock()
	level := l.level
	l.mu.RUnlock()

	if level <= LogLevelWarning {
		l.warnLog.Printf(format, args...)
		l.checkRotation()
	}
}

// Error 记录错误信息
func (l *Logger) Error(format string, args ...interface{}) {
	l.mu.RLock()
	level := l.level
	l.mu.RUnlock()

	if level <= LogLevelError {
		l.errorLog.Printf(format, args...)
		l.checkRotation()
	}
}

// Success 记录成功信息（带 ✓ 标记）
func (l *Logger) Success(format string, args ...interface{}) {
	l.mu.RLock()
	level := l.level
	l.mu.RUnlock()

	if level <= LogLevelInfo {
		msg := fmt.Sprintf("✓ "+format, args...)
		l.infoLog.Println(msg)
		l.checkRotation()
	}
}

// Tip 记录提示信息（带 💡 标记）
func (l *Logger) Tip(format string, args ...interface{}) {
	l.mu.RLock()
	level := l.level
	l.mu.RUnlock()

	if level <= LogLevelInfo {
		msg := fmt.Sprintf("💡 "+format, args...)
		l.infoLog.Println(msg)
		l.checkRotation()
	}
}

// checkRotation 检查是否需要轮换日志文件
func (l *Logger) checkRotation() {
	if l.logFile == nil {
		return
	}

	// 获取当前文件大小
	fileInfo, err := l.logFile.Stat()
	if err != nil {
		return
	}

	if fileInfo.Size() >= l.maxSize {
		l.mu.Lock()
		_ = l.rotateNoLock() //nolint:errcheck // rotation failure doesn't stop logging
		l.mu.Unlock()
	}
}

// Close 关闭日志记录器
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// LogLevelFromString 从字符串解析日志级别
func LogLevelFromString(level string) LogLevel {
	switch level {
	case "debug", "DEBUG":
		return LogLevelDebug
	case "info", "INFO":
		return LogLevelInfo
	case "warning", "warn", "WARNING", "WARN":
		return LogLevelWarning
	case "error", "ERROR":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarning:
		return "WARNING"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}
