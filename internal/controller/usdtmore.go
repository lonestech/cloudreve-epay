package controller

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// USDTMorePurchase 处理 USDTMore 支付请求
func (pc *CloudrevePayController) USDTMorePurchase(c *gin.Context) {
	orderId := c.Param("id")
	if orderId == "" {
		logrus.Debugln("无效的订单号")
		c.HTML(http.StatusOK, "error.tmpl", gin.H{
			"message": "无效的订单号",
		})
		return
	}

	// 获取订单信息
	req, ok := pc.Cache.Get(PurchaseSessionPrefix + orderId)
	if !ok {
		logrus.WithField("id", orderId).Debugln("订单信息不存在")
		c.HTML(http.StatusOK, "error.tmpl", gin.H{
			"message": "订单信息不存在",
		})
		return
	}

	order, ok := req.(*PurchaseRequest)
	if !ok {
		logrus.WithField("id", orderId).Debugln("订单信息非法")
		c.HTML(http.StatusOK, "error.tmpl", gin.H{
			"message": "订单信息非法",
		})
		return
	}

	// 计算金额（分转元）
	amount := decimal.NewFromInt(int64(order.Amount)).Div(decimal.NewFromInt(100)).InexactFloat64()

	// 构建回调 URL
	baseURL, _ := url.Parse(pc.Conf.Base)
	callbackURL, _ := url.Parse("/api/v4/callback/custom/" + order.OrderNo)
	returnURL, _ := url.Parse("/return/" + order.OrderNo)

	// 直接创建 USDT 支付订单，不依赖外部 USDTMore API
	logrus.WithFields(logrus.Fields{
		"OrderID":     order.OrderNo,
		"Name":        order.Name,
		"Amount":      amount,
		"NotifyURL":   baseURL.ResolveReference(callbackURL).String(),
		"RedirectURL": baseURL.ResolveReference(returnURL).String(),
	}).Info("USDTMore 支付请求参数")

	// 使用模拟数据创建交易
	tradeID := order.OrderNo + "_" + fmt.Sprintf("%d", time.Now().Unix())
	
	// 获取 USDT 地址（这里使用模拟地址，实际应用中应该使用真实地址）
	usdtAddress := "TRX7YHbYPJYCJtSzKQtLLqZ5hEyLpbsXZH"
	
	// 计算实际支付金额（这里简单地使用原始金额）
	actualAmount := fmt.Sprintf("%.2f", amount)
	
	// 计算过期时间
	expirationTime := 3600 // 1小时
	expirationTimeFormatted := time.Now().Add(time.Duration(expirationTime) * time.Second).Format("2006-01-02 15:04:05")
	
	// 生成支付 URL
	paymentURL := baseURL.String() + "/pay/checkout-counter/" + tradeID
	
	// 生成二维码 URL
	qrCodeURL := fmt.Sprintf("https://chart.googleapis.com/chart?chs=200x200&cht=qr&chl=%s&choe=UTF-8", 
		url.QueryEscape(fmt.Sprintf("tether:%s?amount=%s", usdtAddress, actualAmount)))

	logrus.WithFields(logrus.Fields{
		"TradeID": tradeID,
		"OrderID": order.OrderNo,
		"Amount": amount,
		"ActualAmount": actualAmount,
		"Address": usdtAddress,
		"ExpirationTime": expirationTimeFormatted,
		"PaymentURL": paymentURL,
		"QRCodeURL": qrCodeURL,
	}).Info("USDTMore 交易创建成功")

	// 渲染 USDT 支付页面
	c.HTML(http.StatusOK, "usdtmore.tmpl", gin.H{
		"OrderId":     order.OrderNo,
		"Name":        order.Name,
		"Amount":      strconv.FormatFloat(amount, 'f', 2, 64),
		"USDTAmount":  actualAmount,
		"Address":     usdtAddress,
		"QRCodeURL":   qrCodeURL,
		"ExpireTime":  expirationTimeFormatted,
		"PaymentURL":  paymentURL,
		"RedirectURL": baseURL.ResolveReference(returnURL).String(),
	})
}

// USDTMoreCheckStatus 检查 USDTMore 订单状态
func (pc *CloudrevePayController) USDTMoreCheckStatus(c *gin.Context) {
	tradeId := c.Param("trade_id")
	if tradeId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "无效的交易ID",
		})
		return
	}

	// 解析 tradeId 获取订单号
	// 格式：orderNo_timestamp
	parts := strings.Split(tradeId, "_")
	if len(parts) < 2 {
		logrus.WithField("trade_id", tradeId).Warn("无效的交易ID格式")
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "无效的交易ID格式",
		})
		return
	}

	// 在实际应用中，这里应该查询数据库检查订单状态
	// 为了简化演示，我们假设所有订单都是等待支付状态
	// 在生产环境中，应该实现真正的订单状态检查逻辑
	status := "waiting"

	// 记录检查结果
	logrus.WithFields(logrus.Fields{
		"trade_id": tradeId,
		"status":   status,
	}).Debug("检查 USDTMore 订单状态")

	c.JSON(http.StatusOK, gin.H{
		"code":   0,
		"status": status,
	})
}
