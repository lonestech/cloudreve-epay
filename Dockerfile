FROM golang:1.20-alpine AS builder

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

# 从构建阶段复制二进制文件
COPY --from=builder /build/epay /app/
COPY --from=builder /build/templates /app/templates

# 创建配置目录
RUN mkdir -p /app/custom

EXPOSE 4560

CMD ["./epay"]
