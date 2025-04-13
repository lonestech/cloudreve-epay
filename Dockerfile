FROM golang:1.24-alpine AS builder

WORKDIR /build

# 安装必要的构建工具
RUN apk add --no-cache gcc musl-dev

# 复制源代码
COPY . .

# 构建应用
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o epay .

FROM alpine:latest

WORKDIR /app

# 安装基本依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 从构建阶段复制二进制文件和模板
COPY --from=builder /build/epay /app/

# 创建配置目录并复制模板
RUN mkdir -p /app/custom/templates
COPY --from=builder /build/templates/*.tmpl /app/custom/templates/

# 确保模板文件存在于正确的位置
RUN ls -la /app/custom/templates/*.tmpl || (echo "Template files are missing" && exit 1)

# 环境变量已在 docker-compose.yml 中定义，不需要复制 .env 文件

EXPOSE 4560

# 添加启动命令
CMD ["/app/epay"]
