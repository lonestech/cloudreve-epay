package usdtmore

import (
	"github.com/imroc/req/v3"
	"go.uber.org/fx"
)

// Module USDTMore 模块
var Module = fx.Options(
	fx.Provide(
		NewConfigFromEnv,
		ProvideClient,
	),
)

// NewConfigFromEnv 从环境变量创建配置
func NewConfigFromEnv() *Config {
	// 这里实际应该从环境变量读取配置
	// 暂时使用硬编码的配置，后续会修改为从环境变量读取
	return &Config{
		Enabled:      true,
		APIEndpoint:  "http://localhost:6080",
		AuthToken:    "123456",
		DefaultChain: "TRON",
	}
}

// ProvideClient 提供 USDTMore 客户端
func ProvideClient(config *Config, client *req.Client) *Client {
	return &Client{
		config: config,
		client: client,
	}
}
