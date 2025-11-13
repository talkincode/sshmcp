# Git Hooks

This directory contains Git hooks to ensure code quality and consistency.

## Available Hooks

### pre-commit

Runs before each commit to ensure:

- ✅ Code is properly formatted (`go fmt`)
- ✅ No vet issues (`go vet`)
- ✅ All tests pass (`go test`)
- ✅ Build is successful (`go build`)
- ✅ Linting passes (`golangci-lint` if installed)

### commit-msg

Enforces conventional commit message format:

```
<type>(<scope>): <subject>

Types: feat, fix, docs, style, refactor, perf, test, chore, build, ci
```

**Examples:**

```bash
git commit -m "feat: 添加 HTTP 代理支持"
git commit -m "fix(ssh): 修复连接超时问题"
git commit -m "docs: 更新 README 安装说明"
```

### pre-push

Runs before pushing to ensure:

- ✅ All tests pass with race detection
- ✅ Test coverage ≥ 50%
- ✅ Cross-platform builds succeed

## Installation

Run the setup script:

```bash
./scripts/setup-hooks.sh
```

Or manually configure:

```bash
git config core.hooksPath .githooks
```

## Bypassing Hooks

**Not recommended, but sometimes necessary:**

```bash
# Skip pre-commit and commit-msg hooks
git commit --no-verify -m "message"

# Skip pre-push hook
git push --no-verify
```

## Testing Hooks

Test a hook manually before committing:

```bash
# Test pre-commit hook
./.githooks/pre-commit

# Test commit-msg hook
echo "feat: test message" > /tmp/msg
./.githooks/commit-msg /tmp/msg
```

## Requirements

### Required

- Go 1.21+
- Git 2.9+

### Optional (but recommended)

- [golangci-lint](https://golangci-lint.run/) - Enhanced linting

  ```bash
  # macOS
  brew install golangci-lint

  # Linux
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
  ```

## Troubleshooting

### Hooks not running

Check if hooks are configured:

```bash
git config core.hooksPath
```

Should output: `.githooks`

### Permission denied

Make hooks executable:

```bash
chmod +x .githooks/*
```

### Tests failing

Run tests locally to see details:

```bash
make test
# or
go test -v ./...
```

### Slow pre-commit

If hooks are too slow, consider:

1. Running only fast tests in pre-commit
2. Moving comprehensive tests to pre-push
3. Using `--no-verify` for WIP commits (not recommended)

## Customization

Edit hooks in `.githooks/` directory. Changes take effect immediately.

## CI Integration

These hooks mirror the checks in `.github/workflows/ci.yml`:

- **Local (hooks)**: Fast feedback before commit/push
- **Remote (CI)**: Comprehensive checks on multiple platforms

Both must pass for code to be merged.
