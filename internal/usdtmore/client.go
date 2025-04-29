package usdtmore

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
)

// Client USDTMore 客户端
type Client struct {
	config *Config
	client *req.Client
}

// NewClient 创建 USDTMore 客户端
func NewClient(config *Config, client *req.Client) *Client {
	return &Client{
		config: config,
		client: client,
	}
}

// CreateTransactionResponse USDTMore 创建交易响应
type CreateTransactionResponse struct {
	TradeID       string  `json:"trade_id"`
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	ActualAmount  string  `json:"actual_amount"`
	Token         string  `json:"token"`
	ExpirationTime int    `json:"expiration_time"`
	PaymentURL    string  `json:"payment_url"`
}

// CreateTransaction 创建交易
func (c *Client) CreateTransaction(orderID string, amount float64, notifyURL, redirectURL string) (*CreateTransactionResponse, error) {
	// 构建请求数据
	data := map[string]interface{}{
		"order_id":     orderID,
		"amount":       amount,
		"notify_url":   notifyURL,
		"redirect_url": redirectURL,
		"code":         c.config.DefaultChain,
	}

	// 生成签名
	signature := c.generateSignature(data)
	data["signature"] = signature

	logrus.WithField("data", data).Debug("创建 USDTMore 交易请求")

	// 发送请求
	url := c.config.APIEndpoint + "/api/v1/order/create-transaction"
	logrus.WithField("url", url).Debug("发送请求到 USDTMore API")

	resp, err := c.client.R().
		SetBody(data).
		Post(url)

	if err != nil {
		logrus.WithError(err).Error("发送请求到 USDTMore API 失败")
		return nil, fmt.Errorf("发送请求到 USDTMore API 失败: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"status": resp.StatusCode,
		"body": string(resp.Bytes()),
	}).Debug("USDTMore API 响应")

	var result struct {
		Code int                     `json:"code"`
		Msg  string                  `json:"msg"`
		Data *CreateTransactionResponse `json:"data"`
	}

	err = json.Unmarshal(resp.Bytes(), &result)
	if err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("USDTMore API 错误: %s", result.Msg)
	}

	return result.Data, nil
}

// CheckOrderStatus 检查订单状态
func (c *Client) CheckOrderStatus(tradeID string) (string, error) {
	resp, err := c.client.R().
		Get(c.config.APIEndpoint + "/pay/check-status/" + tradeID)

	if err != nil {
		return "", err
	}

	var result struct {
		TradeID   string `json:"trade_id"`
		Status    int    `json:"status"`
		ReturnURL string `json:"return_url"`
	}

	err = json.Unmarshal(resp.Bytes(), &result)
	if err != nil {
		return "", err
	}

	// 状态：0-等待支付，1-支付成功，2-已过期
	switch result.Status {
	case 0:
		return "waiting", nil
	case 1:
		return "success", nil
	case 2:
		return "expired", nil
	default:
		return "unknown", nil
	}
}

// generateSignature 生成签名
func (c *Client) generateSignature(data map[string]interface{}) string {
	// 按照键名排序
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建签名字符串
	var signStr strings.Builder
	for _, k := range keys {
		signStr.WriteString(k)
		signStr.WriteString("=")
		signStr.WriteString(fmt.Sprintf("%v", data[k]))
		signStr.WriteString("&")
	}
	signStr.WriteString("token=")
	signStr.WriteString(c.config.AuthToken)

	// 计算 MD5
	hash := md5.Sum([]byte(signStr.String()))
	return fmt.Sprintf("%x", hash)
}
