# 五常同城 - 后端 Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git gcc musl-dev

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server/main.go

# 运行阶段
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata curl && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/migrations ./migrations

RUN mkdir -p uploads logs

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

CMD ["./server"]