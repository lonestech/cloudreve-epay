package controller

import (
	"encoding/gob"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/topjohncian/cloudreve-pro-epay/internal/epay"
)

const (
	paymentTTL            = 3600 * 24 // 24h
	PurchaseSessionPrefix = "purchase_session_"
)

type PurchaseRequest struct {
	Name      string `json:"name" binding:"required"`
	OrderNo   string `json:"order_no" binding:"required"`
	NotifyUrl string `json:"notify_url" binding:"required"`
	Amount    int    `json:"amount" binding:"required"` // 订单金额，使用货币的最小单位
	Currency  string `json:"currency" binding:"required"` // 货币的 ISO 4217 代码
}

type PurchaseResponse struct {
	Code int    `json:"code"`
	Data string `json:"data"`
}

func init() {
	gob.Register(&PurchaseRequest{})
}

func (pc *CloudrevePayController) Purchase(c *gin.Context) {
	var req PurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Debugln("无法解析请求")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"data":    "",
			"message": "无法解析请求" + err.Error(),
		})
		return
	}

	// 验证货币类型
	if req.Currency != "CNY" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"data":    "",
			"message": "不支持的货币类型：" + req.Currency,
		})
		return
	}

	// 验证金额范围（0.01-1000000.00）
	if req.Amount < 1 || req.Amount > 100000000 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"data":    "",
			"message": "金额超出范围",
		})
		return
	}

	// 验证回调 URL 格式
	if _, err := url.Parse(req.NotifyUrl); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"data":    "",
			"message": "无效的回调 URL",
		})
		return
	}

	if err := pc.Cache.Set(PurchaseSessionPrefix+req.OrderNo, req, paymentTTL); err != nil {
		logrus.WithError(err).Warningln("无法保存订单信息")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    "",
			"message": "无法保存订单信息" + err.Error(),
		})
		return
	}

	baseURL, _ := url.Parse(pc.Conf.Base)
	purchaseURL, err := url.Parse("/purchase/" + req.OrderNo)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    "",
			"message": "无法解析 URL" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PurchaseResponse{
		Code: 0,
		Data: baseURL.ResolveReference(purchaseURL).String(),
	})
}

func (pc *CloudrevePayController) PurchasePage(c *gin.Context) {
	orderId := c.Param("id")
	if orderId == "" {
		logrus.Debugln("无效的订单号")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"data":    "",
			"message": "无效的订单号",
		})
		return
	}

	req, ok := pc.Cache.Get(PurchaseSessionPrefix + orderId)
	if !ok {
		logrus.WithField("id", orderId).Debugln("订单信息不存在")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"data":    "",
			"message": "订单信息不存在",
		})
		return
	}

	order, ok := req.(*PurchaseRequest)
	if !ok {
		logrus.WithField("id", orderId).Debugln("订单信息非法")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"data":    "",
			"message": "订单信息非法",
		})
		return
	}

	baseURL, _ := url.Parse(pc.Conf.Base)
	purchaseURL, _ := url.Parse("/notify/" + order.OrderNo)
	returnURL, err := url.Parse("/return/" + order.OrderNo)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    "",
			"message": "无法解析 URL" + err.Error(),
		})
		return
	}

	// 将订单金额从分转换为元（除以100），并保留两位小数
	amount := decimal.NewFromInt(int64(order.Amount)).Div(decimal.NewFromInt(100)).StringFixedBank(2)

	// 获取客户端IP
	clientIP := c.ClientIP()

	// 判断设备类型
	deviceType := epay.PC
	userAgent := c.Request.UserAgent()
	if strings.Contains(strings.ToLower(userAgent), "mobile") {
		deviceType = epay.MOBILE
	}

	args := &epay.PurchaseArgs{
		Type:           epay.PurchaseType(pc.Conf.EpayPurchaseType),
		ServiceTradeNo: order.OrderNo,
		Name:           order.Name,
		Money:          amount,
		Device:         deviceType,
		ClientIP:       clientIP,
		NotifyUrl:      baseURL.ResolveReference(purchaseURL),
		ReturnUrl:      baseURL.ResolveReference(returnURL),
	}

	if pc.Conf.CustomName != "" {
		args.Name = pc.Conf.CustomName
	}

	client := epay.NewClient(&epay.Config{
		PartnerID: pc.Conf.EpayPartnerID,
		Key:       pc.Conf.EpayKey,
		Endpoint:  pc.Conf.EpayEndpoint,
	})

	endpoint, purchaseParams := client.Purchase(args)

	c.HTML(http.StatusOK, "payment_template.html", gin.H{
		"Endpoint": endpoint,
		"Params":   purchaseParams,
		"Name":     args.Name,
		"OrderNo":  args.ServiceTradeNo,
		"Money":    args.Money,
		"PayType":  args.Type,
		"AutoSubmit": pc.Conf.AutoSubmit,
	})
}
