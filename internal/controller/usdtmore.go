package controller

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/topjohncian/cloudreve-pro-epay/internal/usdtmore"
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

	// 创建 USDTMore 交易
	client := pc.USDTMoreClient
	if client == nil {
		logrus.Errorln("USDTMore 客户端未初始化")
		c.HTML(http.StatusOK, "error.tmpl", gin.H{
			"message": "USDT 支付服务当前不可用",
		})
		return
	}

	// 记录请求参数
	logrus.WithFields(logrus.Fields{
		"OrderID":     order.OrderNo,
		"Name":        order.Name,
		"Amount":      amount,
		"NotifyURL":   baseURL.ResolveReference(callbackURL).String(),
		"RedirectURL": baseURL.ResolveReference(returnURL).String(),
	}).Info("USDTMore 支付请求参数")

	result, err := client.Purchase(&usdtmore.PurchaseArgs{
		OrderID:     order.OrderNo,
		Name:        order.Name,
		Amount:      amount,
		NotifyURL:   baseURL.ResolveReference(callbackURL).String(),
		RedirectURL: baseURL.ResolveReference(returnURL).String(),
	})

	if err != nil {
		logrus.WithError(err).Errorln("创建 USDTMore 交易失败")
		c.HTML(http.StatusOK, "error.tmpl", gin.H{
			"message": "创建 USDT 支付订单失败: " + err.Error(),
		})
		return
	}

	// 渲染 USDT 支付页面
	c.HTML(http.StatusOK, "usdtmore.tmpl", gin.H{
		"OrderId":     order.OrderNo,
		"Name":        order.Name,
		"Amount":      strconv.FormatFloat(amount, 'f', 2, 64),
		"USDTAmount":  result.ActualAmount,
		"Address":     result.Token,
		"QRCodeURL":   result.GetQRCodeURL(),
		"ExpireTime":  result.FormatExpirationTime(),
		"PaymentURL":  result.PaymentURL,
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

	// 检查订单状态
	client := pc.USDTMoreClient
	status, err := client.CheckOrderStatus(tradeId)
	if err != nil {
		logrus.WithError(err).Errorln("检查 USDTMore 订单状态失败")
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  "检查订单状态失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   0,
		"status": status,
	})
}
