# 多阶段构建：前端 Vue 应用 + Nginx
# syntax=docker/dockerfile:1

# ====== 阶段1：构建 ======
FROM node:20-alpine AS builder

WORKDIR /build

# 先拷贝依赖文件，利用缓存层
COPY package.json package-lock.json ./
RUN npm ci --no-audit --no-fund

# 拷贝源码与配置
COPY . .

# 构建生产包
RUN npm run build

# ====== 阶段2：运行（Nginx 托管静态资源 + 反代 API） ======
FROM nginx:1.27-alpine

# 时区
RUN apk add --no-cache tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 拷贝自定义 Nginx 配置
COPY deploy/nginx.conf /etc/nginx/conf.d/default.conf

# 拷贝构建产物
COPY --from=builder /build/dist /usr/share/nginx/html

ENV TZ=Asia/Shanghai

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
