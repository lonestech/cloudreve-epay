# Cloudreve 易支付网关

适配 Cloudreve Pro 的易支付网关，支持支付宝、微信支付等多种支付方式，提供美观的支付界面和灵活的配置选项。

## ✨ 特性

- 🎨 美观的支付界面，支持自定义模板
- 🔄 支持支付宝、微信支付等多种支付方式
- 📱 完美适配移动端和桌面端
- 🔒 安全的支付流程，支持 Redis 持久化
- ⚡ 简单易用的配置系统

## 📋 系统要求

- Cloudreve Pro 3.7.1 及以上版本
- 建议使用 Redis 持久化支付数据

## 📍 快速开始

### 方式一：直接运行

```bash
# 1. 下载最新版本
从 Releases 页面下载对应系统和架构的可执行文件

# 2. 配置环境
复制 .env.example 到 .env
根据实际情况修改配置

# 3. 启动服务
./cloudreve-epay
```

### 方式二：Docker 部署

本项目提供了 Docker Compose 配置，可以一键部署支付服务和 Redis：

1. 准备配置文件
```bash
# 复制示例配置
cp .env.example .env

# 修改必要的配置项
CR_EPAY_BASE=https://payment.example.com        # 站点外部访问 URL
CR_EPAY_CLOUDREVE_KEY=your-secret-key          # Cloudreve 通信密钥
CR_EPAY_EPAY_PARTNER_ID=your-partner-id        # 易支付商户 ID
CR_EPAY_EPAY_KEY=your-epay-key                 # 易支付商户密钥
CR_EPAY_EPAY_ENDPOINT=https://pay.example.com   # 易支付接口地址
```

2. 启动服务
```bash
# 构建并启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

3. 服务说明
- 支付服务：
  - 端口：4560
  - 自动重启：是
  - 环境变量：从 .env 文件加载
- Redis 服务：
  - 持久化：启用 AOF
  - 数据目录：使用 Docker volume
  - 网络：仅内部访问

### 2. 配置参数

```env
# 基本配置
CR_EPAY_DEBUG=false                            # 是否启用调试模式
CR_EPAY_LISTEN=:4560                           # 监听端口
CR_EPAY_BASE=https://payment.example.com        # 站点外部访问 URL
CR_EPAY_CLOUDREVE_KEY=your-secret-key          # Cloudreve 通信密钥

# 易支付配置
CR_EPAY_EPAY_PARTNER_ID=1010                   # 商户 ID
CR_EPAY_EPAY_KEY=your-epay-key                 # 商户密钥
CR_EPAY_EPAY_ENDPOINT=https://pay.example.com   # 易支付网关
CR_EPAY_EPAY_PURCHASE_TYPE=alipay              # 默认支付方式

# Redis 配置
CR_EPAY_REDIS_ENABLED=true                     # 强烈建议启用
CR_EPAY_REDIS_SERVER=localhost:6379
CR_EPAY_REDIS_PASSWORD=
CR_EPAY_REDIS_DB=0

# 界面配置
CR_EPAY_CUSTOM_NAME=                           # 自定义商品名称
CR_EPAY_PAYMENT_TEMPLATE=payment_template.html  # 支付页面模板
CR_EPAY_AUTO_SUBMIT=true                       # 是否自动跳转
```

## 🔧 Cloudreve 配置

1. 进入 Cloudreve 后台
2. 进入 `参数设置` > `增值服务`
3. 开启 `自定义付款渠道`
4. 配置以下项目：
   - 付款方式名称：自定义
   - 通讯密钥：与 `CR_EPAY_CLOUDREVE_KEY` 保持一致
   - 支付接口地址：`CR_EPAY_BASE` + `/cloudreve/purchase`
5. 保存设置

## 🎨 自定义支付界面

### 模板变量

| 变量 | 说明 |
|---------|--------|
| `{{.Name}}` | 商品名称 |
| `{{.OrderNo}}` | 订单号 |
| `{{.Money}}` | 支付金额 |
| `{{.PayType}}` | 支付方式 |
| `{{.Endpoint}}` | 支付网关 |
| `{{.Params}}` | 支付参数 |
| `{{.AutoSubmit}}` | 自动提交 |

### 使用方式

1. 直接修改默认模板 `payment_template.html`
2. 或创建新模板并配置 `CR_EPAY_PAYMENT_TEMPLATE`

## 📈 更新日志

### v0.3.0

- ✨ 新增美观的支付界面
- 🎨 支持自定义模板
- 🔄 支持自动/手动跳转支付
- 📱 优化移动端体验

### v0.2.0

- ✅ 支持多种支付方式
- ✨ 支持自定义商品名称
