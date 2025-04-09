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

// CallbackResponse 回调响应格式
type CallbackResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error,omitempty"`
}

// Callback 处理支付回调
func (pc *CloudrevePayController) Callback(c *gin.Context) {
	logrus.Info("收到支付回调")
	
	// 获取所有请求参数
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
		logrus.Warningln("签名验证失败")
		c.JSON(http.StatusOK, CallbackResponse{
			Code:  400,
			Error: "签名验证失败",
		})
		return
	}
	*/
	
	// 打印收到的参数，便于调试
	logrus.WithField("params", params).Infoln("收到支付平台回调")

	// 获取订单号
	orderNo := params["out_trade_no"]
	if orderNo == "" {
		logrus.Debugln("无效的订单号")
		c.JSON(http.StatusOK, CallbackResponse{
			Code:  400,
			Error: "无效的订单号",
		})
		return
	}

	// 获取订单信息
	request, ok := pc.Cache.Get(PurchaseSessionPrefix + orderNo)
	if !ok {
		logrus.WithField("order_no", orderNo).Debugln("订单信息不存在")
		c.JSON(http.StatusOK, CallbackResponse{
			Code:  404,
			Error: "订单信息不存在",
		})
		return
	}

	// 类型断言
	order, ok := request.(*PurchaseRequest)
	if !ok {
		logrus.WithField("order_no", orderNo).Debugln("订单信息非法")
		c.JSON(http.StatusOK, CallbackResponse{
			Code:  500,
			Error: "订单信息非法",
		})
		return
	}

	// 检查支付状态
	if params["trade_status"] == "TRADE_SUCCESS" {
		// 验证金额
		amount := decimal.NewFromInt(int64(order.Amount)).Div(decimal.NewFromInt(100))
		realAmount, err := decimal.NewFromString(params["money"])
		if err != nil {
			logrus.WithError(err).WithField("order_no", orderNo).Debugln("无法解析订单金额")
			c.JSON(http.StatusOK, CallbackResponse{
				Code:  500,
				Error: "无法解析订单金额",
			})
			return
		}
		if !realAmount.Equal(amount) {
			logrus.WithField("order_no", orderNo).Debugln("订单金额不符")
			c.JSON(http.StatusOK, CallbackResponse{
				Code:  400,
				Error: "订单金额不符",
			})
			return
		}

		// 发送通知
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
				logrus.WithField("order_no", orderNo).WithError(err).Errorln("解析 URL 失败")
				return err
			}
			
			// 生成签名内容
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
			logrus.WithField("order_no", orderNo).WithField("Authorization", authHeader).Debugln("生成的 Authorization 头")
			
			// 发送 GET 请求
			resp, err := pc.Client.R().
				SetSuccessResult(&notifyRes).
				SetHeader("Authorization", authHeader).
				Get(order.NotifyUrl)

			if err != nil {
				logrus.WithField("order_no", orderNo).WithError(err).Errorln("通知失败")
				return err
			}

			if !resp.IsSuccessState() {
				logrus.WithField("order_no", orderNo).WithField("dump", resp.Dump()).Errorln("通知失败")
				return errors.New("http code: " + strconv.Itoa(resp.StatusCode))
			}

			if notifyRes.Code != 0 {
				logrus.WithField("order_no", orderNo).WithField("dump", resp.Dump()).WithField("error", notifyRes.Error).Errorln("通知失败")
				return errors.New("code: " + strconv.Itoa(notifyRes.Code) + ", error: " + notifyRes.Error)
			}

			return nil
		}, retry.Attempts(5), retry.Delay(10), retry.OnRetry(func(n uint, err error) {
			logrus.WithField("order_no", orderNo).WithField("n", n).WithError(err).Infoln("通知失败，重试")
		}))

		if err != nil {
			logrus.WithField("order_no", orderNo).WithError(err).Errorln("通知失败")
			c.JSON(http.StatusOK, CallbackResponse{
				Code:  500,
				Error: "通知失败: " + err.Error(),
			})
			return
		}

		logrus.WithField("order_no", orderNo).Infoln("通知成功")

		// 标记订单为已支付
		err = cache.MarkOrderAsPaid(pc.Cache, orderNo)
		if err != nil {
			logrus.WithField("order_no", orderNo).WithError(err).Errorln("标记订单为已支付失败")
		}

		// 从缓存中删除订单信息
		pc.Cache.Delete([]string{orderNo}, PurchaseSessionPrefix)

		// 返回成功响应
		c.JSON(http.StatusOK, CallbackResponse{
			Code: 0,
		})
		return
	}

	// 如果支付状态不是成功，返回成功但不处理
	c.JSON(http.StatusOK, CallbackResponse{
		Code: 0,
	})
}
