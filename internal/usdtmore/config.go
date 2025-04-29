package usdtmore

// Config USDTMore 配置
type Config struct {
	Enabled      bool   // 是否启用 USDTMore 支付
	APIEndpoint  string // USDTMore API 端点
	AuthToken    string // USDTMore 认证令牌
	DefaultChain string // 默认链路：TRON, POLY, OP, BSC
}

// NewConfig 创建 USDTMore 配置
func NewConfig(enabled bool, apiEndpoint, authToken, defaultChain string) *Config {
	return &Config{
		Enabled:      enabled,
		APIEndpoint:  apiEndpoint,
		AuthToken:    authToken,
		DefaultChain: defaultChain,
	}
}
