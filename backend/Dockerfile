# 多阶段构建：后端 Go 服务
# syntax=docker/dockerfile:1

# ====== 阶段1：构建 ======
FROM golang:1.25-alpine AS builder

# 安装必要的构建工具（CGO 依赖等，如需可加 gcc musl-dev）
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# 先拷贝依赖文件，利用缓存层
COPY go.mod go.sum ./
RUN go mod download

# 拷贝源码
COPY . .

# 构建可执行文件
# CGO_ENABLED=0 静态编译，便于 scratch/alpine 运行
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags "-s -w -X 'main.Version=${VERSION}' -X 'main.BuildTime=${BUILD_TIME}' -X 'main.GitCommit=${GIT_COMMIT}'" \
    -o /out/wuchang-tongcheng ./cmd/server/main.go

# ====== 阶段2：运行 ======
FROM alpine:3.20

# 时区与证书
RUN apk add --no-cache tzdata ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 拷贝二进制
COPY --from=builder /out/wuchang-tongcheng /app/wuchang-tongcheng

# 配置与上传目录（通过 volume 挂载覆盖）
RUN mkdir -p /app/configs /app/uploads /app/logs

ENV TZ=Asia/Shanghai
ENV GIN_MODE=release

EXPOSE 8080

ENTRYPOINT ["/app/wuchang-tongcheng"]
