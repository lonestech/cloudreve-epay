package usdtmore

import (
	"github.com/imroc/req/v3"
	"github.com/topjohncian/cloudreve-pro-epay/internal/appconf"
	"go.uber.org/fx"
)

// Module USDTMore 模块
var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewConfigFromEnv,
			fx.ParamTags(`group:"config"`),
		),
		ProvideClient,
	),
)

// NewConfigFromEnv 从环境变量创建配置
func NewConfigFromEnv(conf *appconf.Config) *Config {
	return &Config{
		Enabled:      conf.USDTMoreEnabled,
		APIEndpoint:  conf.USDTMoreAPIEndpoint,
		AuthToken:    conf.USDTMoreAuthToken,
		DefaultChain: conf.USDTMoreDefaultChain,
	}
}

// ProvideClient 提供 USDTMore 客户端
func ProvideClient(config *Config, client *req.Client) *Client {
	return &Client{
		config: config,
		client: client,
	}
}
