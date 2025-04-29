package usdtmore

import (
	"fmt"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

// PurchaseArgs USDTMore 购买参数
type PurchaseArgs struct {
	OrderID     string  // 订单号
	Name        string  // 商品名称
	Amount      float64 // 金额（元）
	NotifyURL   string  // 通知 URL
	RedirectURL string  // 重定向 URL
}

// PurchaseResult USDTMore 购买结果
type PurchaseResult struct {
	PaymentURL    string // 支付 URL
	Token         string // 收款地址
	ActualAmount  string // 实际支付金额
	ExpirationTime int    // 过期时间（秒）
}

// Purchase 创建购买请求
func (c *Client) Purchase(args *PurchaseArgs) (*PurchaseResult, error) {
	if !c.config.Enabled {
		return nil, fmt.Errorf("USDTMore 支付未启用")
	}

	// 创建交易
	resp, err := c.CreateTransaction(
		args.OrderID,
		args.Amount,
		args.NotifyURL,
		args.RedirectURL,
	)

	if err != nil {
		logrus.WithError(err).Error("创建 USDTMore 交易失败")
		return nil, err
	}

	// 返回购买结果
	return &PurchaseResult{
		PaymentURL:    resp.PaymentURL,
		Token:         resp.Token,
		ActualAmount:  resp.ActualAmount,
		ExpirationTime: resp.ExpirationTime,
	}, nil
}

// FormatExpirationTime 格式化过期时间
func (r *PurchaseResult) FormatExpirationTime() string {
	expireTime := time.Now().Add(time.Duration(r.ExpirationTime) * time.Second)
	return expireTime.Format("2006-01-02 15:04:05")
}

// GetQRCodeURL 获取二维码 URL
func (r *PurchaseResult) GetQRCodeURL() string {
	// 构建 USDT 转账 URL
	return fmt.Sprintf("https://chart.googleapis.com/chart?chs=200x200&cht=qr&chl=%s&choe=UTF-8", 
		url.QueryEscape(fmt.Sprintf("tether:%s?amount=%s", r.Token, r.ActualAmount)))
}
