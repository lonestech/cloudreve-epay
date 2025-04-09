# Cloudreve 易支付网关

## 项目介绍

这是一个为 Cloudreve Pro 设计的易支付网关集成项目，允许 Cloudreve Pro 用户通过易支付接口处理支付请求。本项目完全实现了 Cloudreve Pro 自定义支付接口规范，包括订单创建、支付通知和订单状态查询功能。

## 功能特性

- ✅ 完整支持 Cloudreve Pro 自定义支付接口规范
- ✅ 支持订单创建、支付通知和订单状态查询
- ✅ 支持多种支付方式（支付宝、微信支付等）
- ✅ 安全的 HMAC 签名验证机制
- ✅ 支持 Redis 缓存，确保支付状态可靠存储
- ✅ 自定义订单名称
- ✅ 支持模板导出，避免 XSS 风险
- ✅ 支持 Cloudreve V4 回调格式

## 系统要求

- Cloudreve Pro 3.7.1 或更高版本
- Go 1.18 或更高版本（如果从源码构建）
- Redis（强烈推荐，但非必须，只测试了开启的情况）
- Docker（可选，用于容器化部署）

## 快速开始

### 部署方法

#### 方法一：二进制文件部署

1. 从 Releases 下载对应系统和架构的二进制可执行文件
2. 复制 `.env.example` 到 `.env`
3. 根据下方配置说明修改 `.env` 文件
4. 启动程序

```bash
# 直接运行
./cloudreve-epay

# 或者使用 systemd 管理（推荐生产环境）
sudo cp cloudreve-epay.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now cloudreve-epay
```

#### 方法二：Docker 容器部署（推荐）

1. 克隆仓库或下载源码

```bash
git clone https://github.com/topjohncian/cloudreve-epay.git
cd cloudreve-epay
```

2. 修改 `docker-compose.yml` 文件中的环境变量

```yaml
environment:
  - CR_EPAY_DEBUG=false
  - CR_EPAY_LISTEN=:4560
  # 下面的变量需要根据实际情况修改
  - CR_EPAY_BASE=http://your-domain.com
  - CR_EPAY_CLOUDREVE_KEY=your_cloudreve_key_here
  - CR_EPAY_EPAY_PARTNER_ID=your_partner_id_here
  - CR_EPAY_EPAY_KEY=your_epay_key_here
  - CR_EPAY_EPAY_ENDPOINT=https://your-epay-endpoint.com/submit.php
  - CR_EPAY_EPAY_PURCHASE_TYPE=alipay
  # Redis 配置
  - CR_EPAY_REDIS_ENABLED=true
  - CR_EPAY_REDIS_SERVER=redis:6379
  - CR_EPAY_REDIS_PASSWORD=
  - CR_EPAY_REDIS_DB=0
  - CR_EPAY_PAYMENT_TEMPLATE=payment_template.html
  - CR_EPAY_AUTO_SUBMIT=true
```

3. 构建并启动容器

```bash
docker-compose build
docker-compose up -d
```

注意：使用 Docker 部署时，不需要创建 `.env` 文件，所有配置都在 `docker-compose.yml` 文件中定义。

### 配置说明

配置参数可以通过环境变量或 `.env` 文件设置。以下是所有支持的配置项：

#### 二进制部署方式（.env 文件）

```env
# 是否启用 debug 模式（生产环境建议设为 false）
CR_EPAY_DEBUG=false

# 监听地址和端口（TLS 请使用其他服务器进行反代）
CR_EPAY_LISTEN=:4560

# 通信密钥（与 Cloudreve 后台设置相同）
# 建议使用随机生成的 UUID: https://www.uuidgenerator.net/
CR_EPAY_CLOUDREVE_KEY=your_secure_communication_key

# 本站点的外部访问 URL（必须是外部可访问的地址）
CR_EPAY_BASE=https://payment.example.com

# 自定义订单名称（可选）
# CR_EPAY_CUSTOM_NAME=我的商店

# 易支付商家 ID
CR_EPAY_EPAY_PARTNER_ID=1010

# 易支付商家密钥
CR_EPAY_EPAY_KEY=your_epay_secret_key

# 易支付网关地址
CR_EPAY_EPAY_ENDPOINT=https://payment.example.com/submit.php

# 支付方式: wxpay（微信支付）或 alipay（支付宝）
CR_EPAY_EPAY_PURCHASE_TYPE=alipay

# Redis 配置（强烈推荐启用）
CR_EPAY_REDIS_ENABLED=true
CR_EPAY_REDIS_SERVER=localhost:6379
# CR_EPAY_REDIS_PASSWORD=your_redis_password
CR_EPAY_REDIS_DB=0
```

#### Docker 部署方式（docker-compose.yml）

在 `docker-compose.yml` 文件中设置环境变量：

```yaml
environment:
  - CR_EPAY_DEBUG=false
  - CR_EPAY_LISTEN=:4560
  # 下面的变量需要根据实际情况修改
  - CR_EPAY_BASE=http://your-domain.com
  - CR_EPAY_CLOUDREVE_KEY=your_cloudreve_key_here
  - CR_EPAY_EPAY_PARTNER_ID=your_partner_id_here
  - CR_EPAY_EPAY_KEY=your_epay_key_here
  - CR_EPAY_EPAY_ENDPOINT=https://your-epay-endpoint.com/submit.php
  - CR_EPAY_EPAY_PURCHASE_TYPE=alipay
  # Redis 配置
  - CR_EPAY_REDIS_ENABLED=true
  - CR_EPAY_REDIS_SERVER=redis:6379
  - CR_EPAY_REDIS_PASSWORD=
  - CR_EPAY_REDIS_DB=0
  - CR_EPAY_PAYMENT_TEMPLATE=payment_template.html
  - CR_EPAY_AUTO_SUBMIT=true
```

注意：使用 Docker 部署时，不需要创建 `.env` 文件，所有配置都在 `docker-compose.yml` 文件中定义。

### Cloudreve 设置

1. 登录 Cloudreve Pro 后台
2. 进入 `参数设置` -> `增值服务`
3. 开启 `自定义付款渠道`
4. 填写以下信息：
   - `付款方式名称`：自定义名称，如"易支付"
   - `通讯密钥`：与 `.env` 文件中 `CR_EPAY_CLOUDREVE_KEY` 的值相同
   - `支付接口地址`：`CR_EPAY_BASE` 的值 + `/cloudreve/purchase`（例如：`https://payment.example.com/cloudreve/purchase`）
5. 保存设置

## 注意事项

1. **版本兼容性**：确保使用 Cloudreve Pro 3.7.1 或更高版本
2. **Redis 缓存**：强烈建议启用 Redis，否则使用内存缓存时，程序重启将导致支付状态丢失
3. **安全配置**：确保 `CR_EPAY_CLOUDREVE_KEY` 使用强密码，并保持其私密性
4. **模板导出**：使用 `-eject` 参数导出模板，避免 XSS 风险
5. **支付方式**：通过 `CR_EPAY_EPAY_PURCHASE_TYPE` 设置默认支付方式，建议选择有自己收银台的易支付服务

## 开发指南

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/topjohncian/cloudreve-epay.git
cd cloudreve-epay

# 安装依赖
go mod tidy

# 构建
go build -o cloudreve-epay
```

### API 文档

本项目实现了 Cloudreve Pro 自定义支付接口规范，详细 API 文档请参考 [custom.md](custom.md)。

## 更新日志

### 0.3 (2025-04-09)

- 实现订单状态查询功能
- 优化缓存机制，提高支付状态可靠性
- 改进 HMAC 签名验证机制
- 更新依赖库，提高安全性

### 0.2

- 修复易支付自定义付款方式，添加 `CR_EPAY_EPAY_PURCHASE_TYPE` 配置
- 支持自定义商品名称

### 0.1

- 初始版本发布
- 实现基本的易支付集成功能

## 贡献指南

欢迎提交 Issues 和 Pull Requests 来帮助改进这个项目！

## 许可证

本项目采用 MIT 许可证。
