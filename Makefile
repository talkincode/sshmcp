.PHONY: help build test test-verbose test-coverage clean install run fmt vet lint deps

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
BINARY_NAME=sshx
BUILD_DIR=bin
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Go 参数
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/$(BUILD_DIR)
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

help: ## 显示帮助信息
	@echo "可用的 Make 目标:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

build: ## 构建二进制文件
	@echo "开始构建..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(GOBIN)/$(BINARY_NAME) ./cmd/sshx
	@echo "构建完成: $(GOBIN)/$(BINARY_NAME)"

build-all: ## 构建所有平台的二进制文件
	@echo "构建所有平台..."
	@mkdir -p $(BUILD_DIR)
	@echo "构建 Linux (amd64)..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(GOBIN)/$(BINARY_NAME)-linux-amd64 ./cmd/sshx
	@echo "构建 Linux (arm64)..."
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(GOBIN)/$(BINARY_NAME)-linux-arm64 ./cmd/sshx
	@echo "构建 macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(GOBIN)/$(BINARY_NAME)-darwin-amd64 ./cmd/sshx
	@echo "构建 macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(GOBIN)/$(BINARY_NAME)-darwin-arm64 ./cmd/sshx
	@echo "构建 Windows (amd64)..."
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(GOBIN)/$(BINARY_NAME)-windows-amd64.exe ./cmd/sshx
	@echo "所有平台构建完成!"

test: ## 运行所有测试
	@echo "运行测试..."
	$(GOTEST) -v ./...

test-short: ## 运行单元测试（跳过集成测试）
	@echo "运行单元测试..."
	$(GOTEST) -v -short ./...

test-verbose: ## 运行详细测试
	@echo "运行详细测试..."
	$(GOTEST) -v -race ./...

test-coverage: ## 运行测试并生成覆盖率报告
	@echo "运行测试并生成覆盖率..."
	$(GOTEST) -v -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./internal/app/...
	@echo "生成覆盖率报告..."
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)
	@echo ""
	@echo "生成 HTML 覆盖率报告..."
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "覆盖率报告已生成: $(COVERAGE_HTML)"

test-app: ## 只测试 app 包
	@echo "测试 app 包..."
	$(GOTEST) -v ./internal/app/...

test-sshclient: ## 只测试 sshclient 包
	@echo "测试 sshclient 包..."
	$(GOTEST) -v ./internal/sshclient/...

clean: ## 清理构建文件和测试缓存
	@echo "清理..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@echo "清理完成!"

install: build ## 安装到 $GOPATH/bin 和 ~/bin
	@echo "安装到系统..."
	@if [ -n "$(GOPATH)" ] && [ -d "$(GOPATH)/bin" ]; then \
		cp $(GOBIN)/$(BINARY_NAME) $(GOPATH)/bin/; \
		echo "✓ 已安装到 $(GOPATH)/bin/$(BINARY_NAME)"; \
	fi
	@if [ -d ~/bin ]; then \
		cp $(GOBIN)/$(BINARY_NAME) ~/bin/$(BINARY_NAME) && chmod +x ~/bin/$(BINARY_NAME); \
		echo "✓ 已安装到 ~/bin/$(BINARY_NAME)"; \
	fi
	@echo "安装完成! 可以使用 '$(BINARY_NAME)' 命令了"

uninstall: ## 从系统卸载
	@echo "卸载..."
	@if [ -f "$(GOPATH)/bin/$(BINARY_NAME)" ]; then \
		rm -f $(GOPATH)/bin/$(BINARY_NAME); \
		echo "✓ 已从 $(GOPATH)/bin 卸载"; \
	fi
	@if [ -f ~/bin/$(BINARY_NAME) ]; then \
		rm -f ~/bin/$(BINARY_NAME); \
		echo "✓ 已从 ~/bin 卸载"; \
	fi
	@echo "卸载完成!"

run: build ## 构建并运行（显示帮助）
	@echo "运行 $(BINARY_NAME)..."
	@$(GOBIN)/$(BINARY_NAME) --help

fmt: ## 格式化代码
	@echo "格式化代码..."
	$(GOFMT) ./...
	@echo "格式化完成!"

vet: ## 运行 go vet 检查
	@echo "运行 go vet..."
	$(GOVET) ./...
	@echo "检查完成!"

lint: ## 运行 golangci-lint (需要先安装)
	@echo "运行 golangci-lint..."
	@which golangci-lint > /dev/null || (echo "请先安装 golangci-lint: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...
	@echo "Lint 检查完成!"

deps: ## 下载依赖
	@echo "下载依赖..."
	$(GOMOD) download
	@echo "依赖下载完成!"

tidy: ## 整理依赖
	@echo "整理依赖..."
	$(GOMOD) tidy
	@echo "依赖整理完成!"

vendor: ## 创建 vendor 目录
	@echo "创建 vendor..."
	$(GOMOD) vendor
	@echo "Vendor 创建完成!"

check: fmt vet test ## 运行所有检查（格式化、vet、测试）
	@echo "所有检查通过!"

ci: deps check test-coverage ## CI/CD 流程（依赖、检查、覆盖率）
	@echo "CI 流程完成!"

tag:
	@echo "🏷️  开始标签创建流程..."
	@./scripts/tag.sh

dev: ## 开发模式（安装依赖、格式化、测试、构建）
	@echo "开发模式..."
	@$(MAKE) deps
	@$(MAKE) fmt
	@$(MAKE) test
	@$(MAKE) build
	@echo "开发环境准备完成!"

release: clean test-coverage build-all ## 发布版本（清理、测试、构建所有平台）
	@echo "准备发布..."
	@echo "所有二进制文件位于: $(BUILD_DIR)/"
	@ls -lh $(BUILD_DIR)/
	@echo "发布准备完成!"

info: ## 显示项目信息
	@echo "项目信息:"
	@echo "  名称: $(BINARY_NAME)"
	@echo "  Go 版本: $(shell go version)"
	@echo "  构建目录: $(BUILD_DIR)"
	@echo "  当前路径: $(GOBASE)"
	@echo ""
	@echo "依赖统计:"
	@go list -m all | wc -l | awk '{print "  总依赖数: " $$1}'
	@echo ""
	@echo "代码统计:"
	@find . -name "*.go" -not -path "./vendor/*" | wc -l | awk '{print "  Go 文件数: " $$1}'
	@find . -name "*_test.go" -not -path "./vendor/*" | wc -l | awk '{print "  测试文件数: " $$1}'

watch: ## 监听文件变化并自动测试（需要安装 entr）
	@echo "监听文件变化..."
	@which entr > /dev/null || (echo "请先安装 entr: brew install entr (macOS) 或 apt-get install entr (Linux)" && exit 1)
	@find . -name "*.go" -not -path "./vendor/*" | entr -c make test

.PHONY: all
all: clean deps fmt vet test build ## 完整构建流程
	@echo "完整构建完成!"
