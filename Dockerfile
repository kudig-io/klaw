# 多阶段构建 - 前端构建阶段
FROM node:18-alpine AS frontend-builder

WORKDIR /app/web

# 复制前端依赖文件
COPY web/package.json web/package-lock.json ./

# 安装依赖
RUN npm ci

# 复制前端源代码
COPY web/ ./

# 构建前端
RUN npm run build

# 后端构建阶段
FROM golang:1.20-alpine AS backend-builder

WORKDIR /app

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o klaw cmd/klaw/main.go

# 最终运行时镜像
FROM alpine:latest

# 安装ca证书（用于Kubernetes API访问）
RUN apk --no-cache add ca-certificates

WORKDIR /app

# 从后端构建阶段复制二进制文件
COPY --from=backend-builder /app/klaw .

# 从前端构建阶段复制构建产物
COPY --from=frontend-builder /app/web/dist ./web/dist

# 复制配置文件
COPY configs/ ./configs/

# 复制技能目录
COPY skills/ ./skills/

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./klaw"]
