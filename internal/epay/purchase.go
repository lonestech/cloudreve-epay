package epay

import (
	"net/url"

	"github.com/mitchellh/mapstructure"
)

type PurchaseType string

var (
	Alipay PurchaseType = "alipay"
	Wxpay  PurchaseType = "wxpay"
)

type DeviceType string

var (
	PC     DeviceType = "pc"
	MOBILE DeviceType = "mobile"
)

type PurchaseArgs struct {
	// 支付类型
	Type PurchaseType
	// 商家订单号
	ServiceTradeNo string
	// 商品名称
	Name string
	// 金额（元），保留两位小数
	Money string
	// 设备类型
	Device DeviceType
	// 客户端IP
	ClientIP string
	// 支付成功后的异步通知地址
	NotifyUrl *url.URL
	// 支付成功后的同步跳转地址
	ReturnUrl *url.URL
}

// const PURCHASE_API = "https://payment.moe/submit.php"

// Purchase 生成支付链接和参数
// 返回支付网关地址和请求参数
func (c *EPayClient) Purchase(args *PurchaseArgs) (string, map[string]string) {
	// 新版易支付API参数
	requestParams := map[string]string{
		"pid":          c.Config.PartnerID,        // 商户ID
		"type":         string(args.Type),         // 支付类型（alipay/wxpay等）
		"out_trade_no": args.ServiceTradeNo,       // 商户订单号
		"notify_url":   args.NotifyUrl.String(),   // 异步通知地址
		"return_url":   args.ReturnUrl.String(),   // 同步跳转地址
		"name":         args.Name,                 // 商品名称
		"money":        args.Money,                // 金额（元），保留两位小数
		"clientip":     args.ClientIP,             // 客户端IP
		"device":       string(args.Device),       // 设备类型（pc/mobile）
		"sign_type":    "MD5",                    // 签名类型
		"sign":         "",                       // 签名（由GenerateParams生成）
	}

	return c.Config.Endpoint, GenerateParams(requestParams, c.Config.Key)
}

const TRADE_SUCCESS = "TRADE_SUCCESS"

type VerifyRes struct {
	// 支付类型
	Type PurchaseType
	// 易支付订单号
	TradeNo string `mapstructure:"trade_no"`
	// 商家订单号
	ServiceTradeNo string `mapstructure:"out_trade_no"`
	// 商品名称
	Name string
	// 金额
	Money string
	// 订单支付状态
	TradeStatus string `mapstructure:"trade_status"`
	// 签名检验
	VerifyStatus bool `mapstructure:"-"`
}

func (c *EPayClient) Verify(params map[string]string) (*VerifyRes, error) {
	sign := params["sign"]
	var verifyRes VerifyRes
	// 从 map 映射到 struct 上
	err := mapstructure.Decode(params, &verifyRes)
	// 验证签名
	verifyRes.VerifyStatus = sign == GenerateParams(params, c.Config.Key)["sign"]
	if err != nil {
		return nil, err
	} else {
		return &verifyRes, nil
	}
}
