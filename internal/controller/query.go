package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/topjohncian/cloudreve-pro-epay/internal/cache"
)

type QueryOrderStatusResponse struct {
	Code  int    `json:"code"`
	Data  string `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// 订单状态常量
const (
	OrderStatusPaid   = "PAID"    // 已支付
	OrderStatusUnpaid = "UNPAID"  // 未支付
)

// QueryOrderStatus handles the GET request to check the payment status of an order
// This implements the specification from custom.md
func (pc *CloudrevePayController) QueryOrderStatus(c *gin.Context) {
	orderNo := c.Query("order_no")
	if orderNo == "" {
		logrus.Debugln("无效的订单号")
		c.JSON(http.StatusOK, QueryOrderStatusResponse{
			Code:  500,
			Error: "Invalid order number",
		})
		return
	}

	// Check if the order is marked as paid first
	if cache.IsOrderPaid(pc.Cache, orderNo) {
		c.JSON(http.StatusOK, QueryOrderStatusResponse{
			Code: 0,
			Data: OrderStatusPaid,
		})
		return
	}

	// Try to get the order from cache
	req, ok := pc.Cache.Get(PurchaseSessionPrefix + orderNo)
	if !ok {
		// If we can't find it in the cache and it's not marked as paid,
		// it's either expired or never existed
		logrus.WithField("order_no", orderNo).Debugln("订单信息不存在")
		c.JSON(http.StatusOK, QueryOrderStatusResponse{
			Code:  0,
			Data:  OrderStatusUnpaid,
		})
		return
	}

	_, ok2 := req.(*PurchaseRequest)
	if !ok2 {
		logrus.WithField("order_no", orderNo).Debugln("订单信息非法")
		c.JSON(http.StatusOK, QueryOrderStatusResponse{
			Code:  500,
			Error: "Invalid order information",
		})
		return
	}

	// 如果订单存在于缓存中，但没有被标记为已支付，则返回未支付状态
	c.JSON(http.StatusOK, QueryOrderStatusResponse{
		Code: 0,
		Data: OrderStatusUnpaid,
	})
}
