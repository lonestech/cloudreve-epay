# Cloudreve 易支付网关

## 项目介绍

这是一个为 Cloudreve Pro 设计的易支付网关集成项目，允许 Cloudreve Pro 用户通过易支付接口处理支付请求。本项目完全实现了 Cloudreve Pro 自定义支付接口规范，包括订单创建、支付通知和订单状态查询功能。

## 功能特性

- ✅ 完整支持 Cloudreve Pro 自定义支付接口规范
- ✅ 支持订单创建、支付通知和订单状态查询
- ✅ 支持多种支付方式（支付宝、微信支付、USDT ）
- ✅ 支持仅启用 USDT 支付模式
- ✅ 安全的 HMAC 签名验证机制
- ✅ 支持 Redis 缓存，确保支付状态可靠存储
- ✅ 自定义订单名称
- ✅ 支持模板导出，避免 XSS 风险
- ✅ 支持 Cloudreve V4 回调格式
- ✅ 支持 USDT 多链支付（TRC20、ERC20、Polygon、BSC ）

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
git clone https://github.com/lonestech/cloudreve-epay.git
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
  - CR_EPAY_AUTO_SUBMIT=true
  # USDTMore 配置（可选，用于支持 USDT 支付）
  - CR_EPAY_USDTMORE_ENABLED=false
  - CR_EPAY_USDTMORE_API_ENDPOINT=http://usdtmore:6080
  - CR_EPAY_USDTMORE_AUTH_TOKEN=your_auth_token
  - CR_EPAY_USDTMORE_DEFAULT_CHAIN=TRON
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

# USDTMore 配置（可选，用于支持 USDT 支付）
CR_EPAY_USDTMORE_ENABLED=false
CR_EPAY_USDTMORE_API_ENDPOINT=http://localhost:6080
CR_EPAY_USDTMORE_AUTH_TOKEN=your_auth_token
CR_EPAY_USDTMORE_DEFAULT_CHAIN=TRON  # 可选值：TRON, POLY, OP, BSC
# 是否只启用USDT支付（如果为true，将只显示USDT支付选项）
CR_EPAY_USDTMORE_ONLY=false
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
  - CR_EPAY_AUTO_SUBMIT=true
  # USDTMore 配置（可选，用于支持 USDT 支付）
  - CR_EPAY_USDT_MORE_ENABLED=false
  - CR_EPAY_USDT_MORE_API_ENDPOINT=http://usdtmore:6080
  - CR_EPAY_USDT_MORE_AUTH_TOKEN=your_auth_token
  # 是否只启用USDT支付（如果为true，将只显示USDT支付选项）
  - CR_EPAY_USDT_MORE_ONLY=false  
  - CR_EPAY_USDT_MORE_DEFAULT_CHAIN=TRON
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

## USDT 支付配置

如果您想启用 USDT 支付功能，需要按照以下步骤进行配置：

1. **部署 USDTMore 服务**：
   - 克隆 USDTMore 仓库：`git clone https://github.com/botinheart/USDTMore.git`
   - 按照 USDTMore 的文档进行部署和配置
   - 确保 USDTMore 服务正常运行，默认端口为 6080

2. **配置 Telegram 机器人**：
   - USDTMore 依赖 Telegram 机器人进行通知和管理
   - 设置 `TG_BOT_TOKEN` 和 `TG_BOT_ADMIN_ID`

3. **仅启用 USDT 支付模式**：
   - 如果您只想提供 USDT 支付选项，可以设置 `CR_EPAY_USDTMORE_ONLY=true`
   - 此配置将使系统只显示 USDT 支付选项，隐藏支付宝/微信支付选项
   - 用户访问支付页面时将直接跳转到 USDT 支付界面   - 详细配置请参考 USDTMore 文档

3. **添加钱包地址**：
   - 通过 Telegram 机器人添加 USDT 收款地址
   - 支持多链地址：TRON(TRC20)、Polygon、Optimism、BSC

4. **配置 cloudreve-epay**：
   - 启用 USDTMore 支持：`CR_EPAY_USDTMORE_ENABLED=true`
   - 设置 API 端点：`CR_EPAY_USDTMORE_API_ENDPOINT=http://your-usdtmore-server:6080`
   - 配置认证令牌：`CR_EPAY_USDTMORE_AUTH_TOKEN=your_auth_token`（与 USDTMore 的 AUTH_TOKEN 保持一致）
   - 选择默认链路：`CR_EPAY_USDTMORE_DEFAULT_CHAIN=TRON`（可选值：TRON, POLY, OP, BSC）

5. **重启 cloudreve-epay 服务**：
   - 重启服务以应用新的配置
   - 在支付页面中将出现 USDT 支付选项

## 反向代理配置

### Caddy

```caddyfile
payment.example.com {
    reverse_proxy localhost:4560
}
```
### Nginx
```
server {
    listen 80;
    server_name payment.example.com;

    location / {
        proxy_pass http://localhost:4560;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 开发指南

### 从源码构建

1. 克隆仓库

```bash
git clone https://github.com/lonestech/cloudreve-epay.git
cd cloudreve-epay
```

2. 安装依赖

```bash
go mod download
```

3. 构建

```bash
go build -o cloudreve-epay main.go
```

### 导出模板

如果您想自定义模板，可以使用 `-eject` 参数导出默认模板：

```bash
./cloudreve-epay -eject
```

这将在当前目录下创建 `custom` 文件夹，您可以修改其中的模板文件。

## 许可证

本项目采用 MIT 许可证。详情请参阅 [LICENSE](LICENSE) 文件。

## 致谢

- [Cloudreve](https://github.com/cloudreve/Cloudreve) - 支持本项目的主要应用
- [Gin](https://github.com/gin-gonic/gin) - HTTP 框架
- [fx](https://github.com/uber-go/fx) - 依赖注入框架
