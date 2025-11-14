# 错误处理和日志机制重构

## 概述

本次重构实现了统一的错误处理和日志记录机制，解决了以下问题：

1. ❌ **旧问题**：错误处理不一致

   - 部分地方使用 `log.Printf` + 忽略错误
   - 部分地方使用 `CloseIgnore` 包装错误
   - EOF 错误的处理逻辑分散在多处

2. ✅ **新方案**：统一的错误处理和日志机制

## 新增包

### 1. `pkg/logger` - 统一日志包

**特性**：

- ✅ 独立包，可被内部和外部使用
- ✅ 默认输出到 `stderr`，不影响 `stdout`
- ✅ 支持文件日志：`~/.sshmcp/sshx.log`
- ✅ 自动日志切分，保留 3 个文件
- ✅ 日志级别：Debug, Info, Warning, Error
- ✅ 线程安全

**使用示例**：

```go
import "github.com/talkincode/sshmcp/pkg/logger"

// 获取全局 logger
lg := logger.GetLogger()

// 记录不同级别的日志
lg.Debug("调试信息: %s", value)
lg.Info("普通信息: %s", value)
lg.Warning("警告信息: %s", value)
lg.Error("错误信息: %s", value)
lg.Success("成功信息: %s", value)  // 带 ✓ 标记
lg.Tip("提示信息: %s", value)     // 带 💡 标记

// 启用文件日志（默认已启用）
if err := lg.EnableFileLogging(""); err != nil {
    // 处理错误
}

// 设置日志级别
lg.SetLevel(logger.LogLevelDebug)

// 设置日志文件大小和数量
lg.SetMaxSize(10 * 1024 * 1024) // 10MB
lg.SetMaxFiles(5)                // 保留 5 个文件
```

### 2. `pkg/errutil` - 统一错误处理包

**特性**：

- ✅ 独立包，可被内部和外部使用
- ✅ 预定义常见错误类型
- ✅ 错误分类：可忽略、可重试、致命
- ✅ 统一的 EOF 错误处理
- ✅ 智能错误增强

**使用示例**：

```go
import "github.com/talkincode/sshmcp/pkg/errutil"

// 检查错误类型
if errutil.IsIgnorableError(err) {
    // 可忽略的错误（如 EOF, net.ErrClosed）
}

if errutil.IsRetriableError(err) {
    // 可重试的错误（如临时网络问题）
}

// 在 defer 中处理 close 错误
func processFile() (err error) {
    file, err := os.Open("file.txt")
    if err != nil {
        return err
    }
    defer errutil.HandleCloseError(&err, file)
    // ... 处理文件 ...
    return nil
}

// 安全关闭资源
if err := errutil.SafeClose(closer); err != nil {
    // 只返回需要关注的错误
}

// 增强错误信息
enhancedErr := errutil.EnhanceError(err, stdout, stderr)
```

## 重构内容

### 1. `internal/sshclient/client.go`

**改进**：

- 统一使用 `logger.GetLogger()` 替代 `log.Printf`
- 统一使用 `errutil.HandleCloseError` 处理 defer 中的错误
- 使用 `errutil.IsEOFError` 统一判断 EOF 错误
- 使用 `errutil.EnhanceError` 增强错误信息

### 2. `internal/sshclient/pool.go`

**改进**：

- 使用 `errutil.SafeClose` 安全关闭连接
- 使用 `logger.GetLogger()` 记录调试和警告信息
- 更好的错误处理和日志记录

### 3. `internal/sshclient/closer.go`

**改进**：

- 增强 `CloseIgnore` 函数，自动识别可忽略错误
- 添加 `MustClose` 便捷函数
- 添加 `SafeCloseMultiple` 批量关闭函数

### 4. `internal/sshclient/validate.go`

**改进**：

- 使用 `logger.GetLogger()` 替代 `log.Printf`

## 日志输出规范

### 控制台输出

- **stderr**：所有日志信息（Debug、Info、Warning、Error）
- **stdout**：命令执行的实际输出（保持干净）

### 文件日志

- **路径**：`~/.sshmcp/sshx.log`
- **切分**：单文件超过 10MB 自动切分
- **保留**：最多保留 3 个文件
  - `sshx.log` - 当前文件
  - `sshx.log.1` - 上一次的文件
  - `sshx.log.2` - 再上一次的文件

## 错误分类

### 可忽略错误（Ignorable）

- `io.EOF` - 文件或连接正常结束
- `net.ErrClosed` - 网络连接已关闭
- `ErrConnectionClosed` - SSH 连接关闭
- `ErrSessionClosed` - SSH 会话关闭

### 可重试错误（Retriable）

- 临时网络错误
- 连接超时
- 连接被拒绝

### 致命错误（Fatal）

- 认证失败
- 配置错误
- 其他不可恢复的错误

## 测试

所有新功能都有完整的单元测试：

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./pkg/logger
go test ./pkg/errutil
go test ./internal/sshclient
```

## 迁移指南

### 从旧代码迁移

**旧代码**：

```go
log.Printf("✓ Connected successfully")

if closeErr := client.Close(); closeErr != nil {
    _ = closeErr  // 忽略错误
}

if err.Error() == "EOF" {
    // 处理 EOF
}
```

**新代码**：

```go
logger.GetLogger().Success("Connected successfully")

_ = errutil.SafeClose(client)

if errutil.IsEOFError(err) {
    // 处理 EOF
}
```

## 优势

1. ✅ **统一性**：所有错误处理和日志记录都使用统一的接口
2. ✅ **可维护性**：集中管理错误处理逻辑
3. ✅ **可扩展性**：独立包可以被其他项目使用
4. ✅ **可测试性**：完整的单元测试覆盖
5. ✅ **用户友好**：stderr 用于日志，stdout 保持干净
6. ✅ **调试友好**：文件日志自动记录，方便问题排查
