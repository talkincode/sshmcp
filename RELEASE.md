# 发布新版本指南

## 自动发布流程

本项目配置了自动化的 CI/CD 流程，当推送新的版本标签时会自动构建并发布。

## 发布步骤

### 1. 更新版本信息

首先更新 `CHANGELOG.md`，记录本次发布的变更：

```markdown
## [1.0.1] - 2025-01-15

### Added
- 新功能说明

### Changed
- 变更说明

### Fixed
- 修复说明
```

### 2. 提交变更

```bash
git add .
git commit -m "chore: prepare for release v1.0.1"
git push origin main
```

### 3. 创建并推送标签

```bash
# 创建标签
git tag -a v1.0.1 -m "Release v1.0.1"

# 推送标签到远程仓库
git push origin v1.0.1
```

### 4. 自动化构建

推送标签后，GitHub Actions 会自动：

1. ✅ 构建以下平台的二进制文件：
   - Linux x86_64
   - Linux ARM64
   - macOS x86_64 (Intel)
   - macOS ARM64 (Apple Silicon)
   - Windows x86_64

2. ✅ 为每个二进制文件创建压缩包：
   - Linux/macOS: `.tar.gz` 格式
   - Windows: `.zip` 格式

3. ✅ 生成 SHA256 校验和文件 (`checksums.txt`)

4. ✅ 自动创建 GitHub Release

5. ✅ 上传所有二进制文件到 Release

### 5. 验证发布

访问 GitHub Releases 页面验证：
```
https://github.com/talkincode/sshmcp/releases
```

检查：
- ✅ Release 已创建
- ✅ 所有 5 个平台的二进制文件已上传
- ✅ checksums.txt 文件存在
- ✅ Release 说明完整

## 版本号规范

遵循 [语义化版本](https://semver.org/lang/zh-CN/) 规范：

- **主版本号 (MAJOR)**: 不兼容的 API 修改
- **次版本号 (MINOR)**: 向下兼容的功能性新增
- **修订号 (PATCH)**: 向下兼容的问题修正

示例：
- `v1.0.0` - 首次正式发布
- `v1.1.0` - 添加新功能
- `v1.1.1` - 修复 bug
- `v2.0.0` - 重大变更，不向下兼容

## 预发布版本

如需发布测试版本：

```bash
# Beta 版本
git tag -a v1.1.0-beta.1 -m "Release v1.1.0-beta.1"
git push origin v1.1.0-beta.1

# Release Candidate 版本
git tag -a v1.1.0-rc.1 -m "Release v1.1.0-rc.1"
git push origin v1.1.0-rc.1
```

预发布版本会在 GitHub Release 中标记为 "Pre-release"。

## 删除错误的标签

如果推送了错误的标签：

```bash
# 删除本地标签
git tag -d v1.0.1

# 删除远程标签
git push origin :refs/tags/v1.0.1
```

## 手动构建（开发测试）

如需在本地测试构建：

```bash
# 构建所有平台
make build-all

# 查看构建结果
ls -lh bin/
```

## 故障排查

### 构建失败

1. 检查 GitHub Actions 日志
2. 验证 `go.mod` 依赖是否正确
3. 确保所有测试通过：`make test`

### Release 未创建

1. 检查标签格式是否为 `v*.*.*`
2. 验证 GitHub Actions 权限设置
3. 检查 `GITHUB_TOKEN` 是否有效

### 文件上传失败

1. 检查文件大小限制
2. 验证网络连接
3. 查看 Actions 日志中的详细错误信息

## CI/CD 工作流

### Release 工作流 (.github/workflows/release.yml)

触发条件：推送 `v*.*.*` 格式的标签

任务：
1. 检出代码
2. 设置 Go 环境
3. 构建多平台二进制
4. 生成校验和
5. 创建 GitHub Release
6. 上传文件

### CI 工作流 (.github/workflows/ci.yml)

触发条件：推送到 main/develop 分支或 Pull Request

任务：
1. 在多个操作系统上运行测试
2. 生成代码覆盖率报告
3. 运行代码检查 (golangci-lint)
4. 运行安全扫描 (gosec)

## 参考资源

- [GitHub Actions 文档](https://docs.github.com/en/actions)
- [Go Release 工作流示例](https://github.com/marketplace/actions/go-release-binaries)
- [语义化版本规范](https://semver.org/lang/zh-CN/)
