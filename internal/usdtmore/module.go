package usdtmore

import (
	"github.com/imroc/req/v3"
	"github.com/topjohncian/cloudreve-pro-epay/internal/appconf"
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
func NewConfigFromEnv(conf *appconf.Config) *Config {
	return NewConfig(
		conf.USDTMoreEnabled,
		conf.USDTMoreAPIEndpoint,
		conf.USDTMoreAuthToken,
		conf.USDTMoreDefaultChain,
	)
}

// ProvideClient 提供 USDTMore 客户端
func ProvideClient(config *Config, client *req.Client) *Client {
	// 如果 USDTMore 未启用，返回 nil
	if !config.Enabled {
		return nil
	}

	// 创建并返回 USDTMore 客户端
	return &Client{
		config: config,
		client: client,
	}
}
