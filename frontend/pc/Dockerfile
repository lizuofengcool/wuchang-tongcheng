# 多阶段构建：构建 Next.js 产物 + 精简运行时
# PC门户站通过 next start 启动（SSR + ISR）

# 阶段1：安装全部依赖（含 devDependencies，build 需要 typescript/tailwind）
FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci || npm install

# 阶段2：构建
FROM node:20-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ENV BACKEND_URL=http://localhost:8080
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build

# 阶段3：精简运行时（仅生产依赖）
FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
RUN addgroup -g 1001 -S nodejs && adduser -S nextjs -u 1001 -G nodejs
COPY --from=builder --chown=nextjs:nodejs /app/.next ./.next
COPY --from=builder --chown=nextjs:nodejs /app/public ./public
COPY --from=deps /app/package.json ./
RUN npm install --omit=dev && chown -R nextjs:nodejs /app/node_modules
USER nextjs
EXPOSE 3000
CMD ["npm", "start"]
