package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/topjohncian/cloudreve-pro-epay/internal/epay"
)

type OrderQueryResponse struct {
	Code    int    `json:"code"`
	Data    string `json:"data"`
	Message string `json:"message"`
}

// QueryOrder 查询订单状态
func (pc *CloudrevePayController) QueryOrder(c *gin.Context) {
	orderNo := c.Query("order_no")
	if orderNo == "" {
		logrus.Debugln("无效的订单号")
		c.AbortWithStatusJSON(http.StatusBadRequest, OrderQueryResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的订单号",
		})
		return
	}

	// 获取订单信息
	request, ok := pc.Cache.Get(PurchaseSessionPrefix + orderNo)
	if !ok {
		logrus.WithField("order_no", orderNo).Debugln("订单信息不存在")
		c.AbortWithStatusJSON(http.StatusNotFound, OrderQueryResponse{
			Code:    http.StatusNotFound,
			Message: "订单信息不存在",
		})
		return
	}

	purchaseReq := request.(PurchaseRequest)

	// 创建易支付客户端
	client := epay.NewClient(&epay.Config{
		PartnerID: pc.Conf.EpayPartnerID,
		Key:       pc.Conf.EpayKey,
		Endpoint:  pc.Conf.EpayEndpoint,
	})

	// 查询订单状态
	verifyRes, err := client.Verify(map[string]string{
		"out_trade_no": orderNo,
	})

	if err != nil {
		logrus.WithField("order_no", orderNo).WithError(err).Errorln("查询订单状态失败")
		c.AbortWithStatusJSON(http.StatusInternalServerError, OrderQueryResponse{
			Code:    http.StatusInternalServerError,
			Message: "查询订单状态失败: " + err.Error(),
		})
		return
	}

	// 返回订单状态
	c.JSON(http.StatusOK, OrderQueryResponse{
		Code: 0,
		Data: verifyRes.TradeStatus,
	})
}
