package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/topjohncian/cloudreve-pro-epay/internal/cache"
)

// CloudreveV4Callback 处理 Cloudreve V4 版本的回调请求
func (pc *CloudrevePayController) CloudreveV4Callback(c *gin.Context) {
	logrus.Info("收到 Cloudreve V4 回调请求")

	// 获取订单号
	orderNo := c.Param("id")
	if orderNo == "" {
		logrus.Debugln("无效的订单号")
		c.JSON(http.StatusOK, gin.H{
			"code":  400,
			"error": "无效的订单号",
		})
		return
	}

	// 记录请求信息，便于调试
	logrus.WithFields(logrus.Fields{
		"order_no": orderNo,
		"method":   c.Request.Method,
		"path":     c.Request.URL.Path,
		"query":    c.Request.URL.Query(),
	}).Infoln("Cloudreve V4 回调请求详情")

	// 获取订单信息
	request, ok := pc.Cache.Get(PurchaseSessionPrefix + orderNo)
	if !ok {
		logrus.WithField("order_no", orderNo).Debugln("订单信息不存在")
		c.JSON(http.StatusOK, gin.H{
			"code":  404,
			"error": "订单信息不存在",
		})
		return
	}

	// 类型断言
	order, ok := request.(*PurchaseRequest)
	if !ok {
		logrus.WithField("order_no", orderNo).Debugln("订单信息非法")
		c.JSON(http.StatusOK, gin.H{
			"code":  500,
			"error": "订单信息非法",
		})
		return
	}

	// 标记订单为已支付
	err := cache.MarkOrderAsPaid(pc.Cache, orderNo)
	if err != nil {
		logrus.WithField("order_no", orderNo).WithError(err).Errorln("标记订单为已支付失败")
		c.JSON(http.StatusOK, gin.H{
			"code":  500,
			"error": "标记订单为已支付失败: " + err.Error(),
		})
		return
	}

	// 从缓存中删除订单信息
	pc.Cache.Delete([]string{orderNo}, PurchaseSessionPrefix)

	logrus.WithField("order_no", orderNo).Infoln("订单已标记为已支付")

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
	})
}
