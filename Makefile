# 五常同城项目 Makefile

# 变量定义
APP_NAME := wuchang-tongcheng
VERSION := 0.1.0
BUILD_TIME := $(shell date "+%Y-%m-%d %H:%M:%S")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_VERSION := $(shell go version)

# Go相关
GO := go
GOFLAGS := -v
LDFLAGS := -ldflags "-X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommit=$(GIT_COMMIT)'"

# 目录
BACKEND_DIR := backend
CMD_DIR := $(BACKEND_DIR)/cmd/server
CONFIG_DIR := $(BACKEND_DIR)/configs
DEPLOY_DIR := deploy

# 默认目标
.PHONY: all
all: build

# 帮助信息
.PHONY: help
help:
	@echo "五常同城项目 Makefile"
	@echo ""
	@echo "用法:"
	@echo "  make <target>"
	@echo ""
	@echo "可用目标:"
	@echo "  all          构建项目（默认）"
	@echo "  build        编译后端服务"
	@echo "  run          运行后端服务"
	@echo "  test         运行测试"
	@echo "  lint         代码检查"
	@echo "  fmt          格式化代码"
	@echo "  tidy         整理依赖"
	@echo "  clean        清理构建产物"
	@echo "  docker-up    启动Docker基础设施"
	@echo "  docker-down  停止Docker基础设施"
	@echo "  config       生成配置文件"
	@echo "  version      显示版本信息"
	@echo "  help         显示帮助信息"

# 版本信息
.PHONY: version
version:
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Go版本: $(GO_VERSION)"

# 编译后端服务
.PHONY: build
build:
	@echo "正在编译后端服务..."
	cd $(BACKEND_DIR) && $(GO) build $(GOFLAGS) $(LDFLAGS) -o bin/$(APP_NAME) $(CMD_DIR)/main.go
	@echo "编译完成，输出: $(BACKEND_DIR)/bin/$(APP_NAME)"

# 运行后端服务
.PHONY: run
run:
	@echo "正在启动后端服务..."
	cd $(BACKEND_DIR) && $(GO) run $(LDFLAGS) $(CMD_DIR)/main.go

# 运行测试
.PHONY: test
test:
	@echo "正在运行测试..."
	cd $(BACKEND_DIR) && $(GO) test -v ./...

# 代码检查
.PHONY: lint
lint:
	@echo "正在进行代码检查..."
	cd $(BACKEND_DIR) && golangci-lint run ./...

# 格式化代码
.PHONY: fmt
fmt:
	@echo "正在格式化代码..."
	cd $(BACKEND_DIR) && $(GO) fmt ./...

# 整理依赖
.PHONY: tidy
tidy:
	@echo "正在整理依赖..."
	cd $(BACKEND_DIR) && $(GO) mod tidy

# 清理构建产物
.PHONY: clean
clean:
	@echo "正在清理构建产物..."
	rm -rf $(BACKEND_DIR)/bin
	rm -rf $(BACKEND_DIR)/dist
	@echo "清理完成"

# 启动Docker基础设施
.PHONY: docker-up
docker-up:
	@echo "正在启动Docker基础设施..."
	cd $(DEPLOY_DIR) && docker-compose up -d
	@echo "Docker基础设施启动完成"

# 停止Docker基础设施
.PHONY: docker-down
docker-down:
	@echo "正在停止Docker基础设施..."
	cd $(DEPLOY_DIR) && docker-compose down
	@echo "Docker基础设施已停止"

# 查看Docker基础设施状态
.PHONY: docker-ps
docker-ps:
	cd $(DEPLOY_DIR) && docker-compose ps

# 查看Docker日志
.PHONY: docker-logs
docker-logs:
	cd $(DEPLOY_DIR) && docker-compose logs -f

# 生成配置文件
.PHONY: config
config:
	@if [ ! -f $(CONFIG_DIR)/config.yaml ]; then \
		echo "正在生成配置文件..."; \
		cp $(CONFIG_DIR)/config.yaml.example $(CONFIG_DIR)/config.yaml; \
		echo "配置文件已生成: $(CONFIG_DIR)/config.yaml"; \
	else \
		echo "配置文件已存在，跳过生成"; \
	fi

# 数据库迁移
# 说明：本项目使用 GORM AutoMigrate，启动服务即自动建表 + seed.Run() 写入种子数据。
# PostGIS 扩展由 deploy/initdb/01-extensions.sql 在首次建库时安装。
# 执行 make migrate 会启动服务一次（完成 AutoMigrate + seed 后可 Ctrl+C 停止 HTTP 服务）。
.PHONY: migrate
migrate:
	@echo "正在执行数据库迁移（AutoMigrate + 种子数据）..."
	@echo "PostGIS 扩展请确保 deploy/initdb/01-extensions.sql 已在首次建库时执行"
	cd $(BACKEND_DIR) && $(GO) run $(CMD_DIR)/main.go

# 生成Swagger文档（需要先安装：go install github.com/swaggo/swag/cmd/swag@latest）
.PHONY: swagger
swagger:
	@echo "正在生成Swagger文档..."
	cd $(BACKEND_DIR) && swag init -g cmd/server/main.go -o docs

# 热重载开发模式（需要air工具）
.PHONY: dev
dev:
	@echo "正在启动开发模式（热重载）..."
	cd $(BACKEND_DIR) && air

# 交叉编译
.PHONY: build-linux
build-linux:
	@echo "正在交叉编译Linux版本..."
	cd $(BACKEND_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o bin/$(APP_NAME)-linux-amd64 $(CMD_DIR)/main.go
	@echo "编译完成，输出: $(BACKEND_DIR)/bin/$(APP_NAME)-linux-amd64"

.PHONY: build-darwin
build-darwin:
	@echo "正在交叉编译macOS版本..."
	cd $(BACKEND_DIR) && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o bin/$(APP_NAME)-darwin-amd64 $(CMD_DIR)/main.go
	@echo "编译完成，输出: $(BACKEND_DIR)/bin/$(APP_NAME)-darwin-amd64"

.PHONY: build-windows
build-windows:
	@echo "正在交叉编译Windows版本..."
	cd $(BACKEND_DIR) && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o bin/$(APP_NAME)-windows-amd64.exe $(CMD_DIR)/main.go
	@echo "编译完成，输出: $(BACKEND_DIR)/bin/$(APP_NAME)-windows-amd64.exe"

# 构建所有平台
.PHONY: build-all
build-all: build-linux build-darwin build-windows
	@echo "所有平台编译完成"
