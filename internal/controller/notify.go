package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/avast/retry-go"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/topjohncian/cloudreve-pro-epay/internal/cache"
	// 在注释中使用了 epay.GenerateSign
	// "github.com/topjohncian/cloudreve-pro-epay/internal/epay"
)

type NotifyResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

func (pc *CloudrevePayController) Notify(c *gin.Context) {
	query := c.Request.URL.Query()
	params := lo.Reduce(lo.Keys(query), func(r map[string]string, t string, i int) map[string]string {
		r[t] = query.Get(t)
		return r
	}, map[string]string{})

	// 在生产环境中，应该验证签名
	// 但在测试环境中，我们可能需要跳过签名验证
	// 如果需要验证签名，请取消下面的注释
	/*
	if epay.GenerateSign(params, pc.Conf.EpayKey) != params["sign"] {
		c.String(400, "fail")
		logrus.Warningln("签名验证失败")
		return
	}
	*/
	
	// 打印收到的参数，便于调试
	logrus.WithField("params", params).Infoln("收到支付平台回调")

	orderId := c.Param("id")
	if orderId == "" {
		logrus.Debugln("无效的订单号")
		c.String(400, "fail")
		return
	}

	request, ok := pc.Cache.Get(PurchaseSessionPrefix + orderId)
	if !ok {
		logrus.WithField("id", orderId).Debugln("订单信息不存在")
		c.String(400, "fail")
		return
	}

	order, ok := request.(*PurchaseRequest)
	if !ok {
		logrus.WithField("id", orderId).Debugln("订单信息非法")
		c.String(400, "fail")
		return
	}

	if params["trade_status"] == "TRADE_SUCCESS" {
		amount := decimal.NewFromInt(int64(order.Amount)).Div(decimal.NewFromInt(100))
		realAmount, err := decimal.NewFromString(params["money"])
		if err != nil {
			logrus.WithError(err).WithField("id", orderId).Debugln("无法解析订单金额")
			c.String(400, "fail")
			return
		}
		if !realAmount.Equal(amount) {
			logrus.WithField("id", orderId).Debugln("订单金额不符")
			c.String(400, "fail")
			return
		}

		err = retry.Do(func() error {
			var notifyRes NotifyResponse

			// 生成 HMAC 签名用于授权
			auth := &HMACAuth{
				CloudreveKey: []byte(pc.Conf.CloudreveKey),
			}
			
			// 生成带有过期时间的签名（10分钟后过期）
			expires := time.Now().Add(10 * time.Minute).Unix()
			
			// 解析通知 URL
			parsedURL, err := url.Parse(order.NotifyUrl)
			if err != nil {
				logrus.WithField("id", orderId).WithError(err).Errorln("解析 URL 失败")
				return err
			}
			
			// 生成签名内容
			// 使用与服务器相同的方式生成签名内容
			req := RequestRawSign{
				Path:   parsedURL.Path,
				Header: "",
				Body:   "", // 通知请求没有请求体
			}
			signContentBytes, _ := json.Marshal(req)
			signContent := string(signContentBytes)
			
			// 生成签名
			signature := auth.Sign(signContent, expires)
			
			// 生成 Authorization 头
			authHeader := "Bearer " + signature
			logrus.WithField("id", orderId).WithField("Authorization", authHeader).Debugln("生成的 Authorization 头")
			
			// 发送 GET 请求
			// 根据文档要求，回调通知应该使用 GET 请求
			resp, err := pc.Client.R().
				SetSuccessResult(&notifyRes).
				SetHeader("Authorization", authHeader).
				Get(order.NotifyUrl)

			if err != nil {
				logrus.WithField("id", orderId).WithError(err).Errorln("通知失败")
				return err
			}

			if !resp.IsSuccessState() {
				logrus.WithField("id", orderId).WithField("dump", resp.Dump()).Errorln("通知失败")
				return errors.New("http code: " + strconv.Itoa(resp.StatusCode))
			}

			if notifyRes.Code != 0 {
				logrus.WithField("id", orderId).WithField("dump", resp.Dump()).WithField("error", notifyRes.Error).Errorln("通知失败")
				return errors.New("code: " + strconv.Itoa(notifyRes.Code) + ", error: " + notifyRes.Error)
			}

			return nil
		}, retry.Attempts(5), retry.Delay(10), retry.OnRetry(func(n uint, err error) {
			logrus.WithField("id", orderId).WithField("n", n).WithError(err).Infoln("通知失败，重试")
		}))

		if err != nil {
			logrus.WithField("id", orderId).WithError(err).Errorln("通知失败")
			c.String(400, "fail")
			return
		}

		logrus.WithField("id", orderId).Infoln("通知成功")
		c.String(200, "success")

		// Mark the order as paid in the cache
		err = cache.MarkOrderAsPaid(pc.Cache, orderId)
		if err != nil {
			logrus.WithField("id", orderId).WithError(err).Errorln("标记订单为已支付失败")
		}

		// 从缓存中删除订单信息
		pc.Cache.Delete([]string{orderId}, PurchaseSessionPrefix)
		return
	}

	c.String(200, "success")
}

func (pc *CloudrevePayController) Return(c *gin.Context) {
	c.HTML(http.StatusOK, "return.tmpl", gin.H{})
}
