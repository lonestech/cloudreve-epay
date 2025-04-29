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
	// 根据独角数卡的实现，在基础 URL 上添加路径
	url := c.config.APIEndpoint + "/api/v1/order/create-transaction"
	logrus.WithField("url", url).Debug("发送请求到 USDTMore API")

	// 发送请求
	// 根据独角数卡的实现，设置 Content-Type 为 application/json
	// 将请求数据转换为 JSON 字符串
	jsonData, err := json.Marshal(data)
	if err != nil {
		logrus.WithError(err).Error("将请求数据转换为 JSON 失败")
		return nil, fmt.Errorf("将请求数据转换为 JSON 失败: %w", err)
	}

	logrus.WithField("json_data", string(jsonData)).Debug("JSON 格式的请求数据")

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(jsonData).
		Post(url)

	if err != nil {
		logrus.WithError(err).Error("发送请求到 USDTMore API 失败")
		return nil, fmt.Errorf("发送请求到 USDTMore API 失败: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"status": resp.StatusCode,
		"body": string(resp.Bytes()),
	}).Debug("USDTMore API 响应")

	// 先打印原始响应内容以便调试
	respContent := string(resp.Bytes())
	logrus.WithField("response_content", respContent).Debug("USDTMore API 原始响应内容")

	// 尝试解析异次元发卡格式的响应
	var acgResult struct {
		StatusCode int                     `json:"status_code"`
		Message    string                  `json:"message"`
		Data       *CreateTransactionResponse `json:"data"`
	}

	err = json.Unmarshal(resp.Bytes(), &acgResult)
	if err != nil {
		logrus.WithError(err).Error("解析 USDTMore API 响应失败")
		return nil, fmt.Errorf("解析 USDTMore API 响应失败: %w", err)
	}

	if acgResult.StatusCode != 200 {
		logrus.WithFields(logrus.Fields{
			"status_code": acgResult.StatusCode,
			"message": acgResult.Message,
		}).Error("USDTMore API 返回错误")
		return nil, fmt.Errorf("USDTMore API 错误: %s", acgResult.Message)
	}

	// 如果数据为空，尝试原始格式
	if acgResult.Data == nil {
		// 尝试原始格式
		var result struct {
			Code int                     `json:"code"`
			Msg  string                  `json:"msg"`
			Data *CreateTransactionResponse `json:"data"`
		}

		err = json.Unmarshal(resp.Bytes(), &result)
		if err != nil {
			logrus.WithError(err).Error("解析 USDTMore API 原始格式响应失败")
			return nil, fmt.Errorf("解析 USDTMore API 原始格式响应失败: %w", err)
		}

		if result.Code != 0 {
			logrus.WithFields(logrus.Fields{
				"code": result.Code,
				"msg": result.Msg,
			}).Error("USDTMore API 返回错误")
			return nil, fmt.Errorf("USDTMore API 错误: %s", result.Msg)
		}

		// 使用原始格式的数据
		acgResult.Data = result.Data
	}

	// 检查响应数据是否为空
	if acgResult.Data == nil {
		logrus.Error("USDTMore API 返回的数据为空")
		return nil, fmt.Errorf("USDTMore API 返回的数据为空")
	}

	logrus.WithFields(logrus.Fields{
		"trade_id": acgResult.Data.TradeID,
		"order_id": acgResult.Data.OrderID,
		"amount": acgResult.Data.Amount,
		"actual_amount": acgResult.Data.ActualAmount,
		"token": acgResult.Data.Token,
		"expiration_time": acgResult.Data.ExpirationTime,
		"payment_url": acgResult.Data.PaymentURL,
	}).Info("USDTMore 交易创建成功")

	return acgResult.Data, nil
}

// CheckOrderStatus 检查订单状态
func (c *Client) CheckOrderStatus(tradeID string) (string, error) {
	// 构建请求 URL
	// 根据独角数卡的实现，在基础 URL 上添加路径
	url := c.config.APIEndpoint + "/api/pay/check-status/" + tradeID
	logrus.WithField("url", url).Debug("发送请求到 USDTMore API 检查订单状态")

	// 发送请求
	// 根据独角数卡的实现，设置 Content-Type 为 application/json
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		Get(url)

	if err != nil {
		logrus.WithError(err).Error("发送请求到 USDTMore API 检查订单状态失败")
		return "", fmt.Errorf("发送请求到 USDTMore API 检查订单状态失败: %w", err)
	}

	// 先打印原始响应内容以便调试
	respContent := string(resp.Bytes())
	logrus.WithField("response_content", respContent).Debug("USDTMore API 检查订单状态原始响应内容")

	// 如果响应为空或者不是 JSON 格式，模拟一个等待支付的状态
	if respContent == "" || strings.HasPrefix(respContent, "<") {
		logrus.Warn("响应不是有效的 JSON，模拟等待支付状态")
		return "waiting", nil
	}

	// 尝试解析响应
	var result struct {
		TradeID   string `json:"trade_id"`
		Status    int    `json:"status"`
		ReturnURL string `json:"return_url"`
	}

	err = json.Unmarshal(resp.Bytes(), &result)
	if err != nil {
		logrus.WithError(err).Error("解析 USDTMore API 检查订单状态响应失败")
		// 如果无法解析，模拟一个等待支付的状态
		return "waiting", nil
	}

	logrus.WithFields(logrus.Fields{
		"trade_id": result.TradeID,
		"status": result.Status,
		"return_url": result.ReturnURL,
	}).Debug("USDTMore 订单状态")

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
	// 根据独角数卡的实现，使用 URL 参数形式拼接
	var signStr strings.Builder
	for i, k := range keys {
		// 跳过空值
		if data[k] == "" {
			continue
		}
		// 跳过 signature 字段
		if k == "signature" {
			continue
		}
		
		// 添加分隔符
		if i > 0 && signStr.Len() > 0 {
			signStr.WriteString("&")
		}
		
		// 拼接为 URL 参数形式
		signStr.WriteString(k)
		signStr.WriteString("=")
		signStr.WriteString(fmt.Sprintf("%v", data[k]))
	}

	// 在签名字符串后面直接添加密钥（不加分隔符）
	signData := signStr.String() + c.config.AuthToken
	
	// 计算 MD5
	hash := md5.Sum([]byte(signData))
	return fmt.Sprintf("%x", hash)
}
