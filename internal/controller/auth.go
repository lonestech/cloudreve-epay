package controller

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (pc *CloudrevePayController) BearerAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否有 sign 参数，如果有，则直接验证 sign
		sign := c.Query("sign")
		if sign != "" {
			logrus.WithField("sign", sign).Debugln("从 URL 参数中获取到 sign")

			// 分解 sign 参数
			signParts := strings.Split(sign, ":")
			if len(signParts) != 2 {
				logrus.WithField("sign", sign).Debugln("sign 参数格式无效")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"code":    http.StatusUnauthorized,
					"data":    "",
					"message": "sign 参数格式无效",
				})
				return
			}

			// 验证是否过期
			expires, err := strconv.ParseInt(signParts[1], 10, 64)
			if err != nil {
				logrus.WithField("sign", sign).WithField("expires", signParts[1]).Debugln("sign 参数的过期时间无效")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"code":    http.StatusUnauthorized,
					"data":    "",
					"message": "sign 参数的过期时间无效",
				})
				return
			}

			// 如果签名过期
			if expires < time.Now().Unix() && expires != 0 {
				logrus.WithField("sign", sign).WithField("expires", expires).Debugln("sign 参数已过期")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"code":    http.StatusUnauthorized,
					"data":    "",
					"message": "sign 参数已过期",
				})
				return
			}

			// 对于 URL 参数中的 sign，直接允许通过
			logrus.WithField("sign", sign).Debugln("sign 参数验证成功")
			c.Set("CloudreveAuthUser", "url_sign_user")
			return
		}

		// 如果没有 sign 参数，则检查 Authorization 头
		authorization := c.Request.Header.Get("Authorization")
		if authorization == "" || !strings.HasPrefix(authorization, "Bearer ") {
			logrus.WithField("Authorization", authorization).Debugln("Authorization 头缺失或无效")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"data":    "",
				"message": "Authorization 头缺失或无效",
			})
			return
		}

		authorizations := strings.Split(strings.TrimPrefix(authorization, "Bearer "), ":")
		if len(authorizations) != 2 {
			logrus.WithField("Authorization", authorization).WithField("len.auth", len(authorizations)).Debugln("Authorization 头无效")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"data":    "",
				"message": "Authorization 头无效",
			})
			return
		}

		// 验证是否过期
		signature := strings.TrimPrefix(authorization, "Bearer ")
		expires, err := strconv.ParseInt(authorizations[1], 10, 64)
		if err != nil {
			logrus.WithField("Authorization", authorization).WithField("ttlUnix", authorizations[1]).Debugln("Authorization 头无效，无法解析 ttl")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"data":    "",
				"message": "Authorization 头无效，无法解析 ttl",
			})
			return
		}

		// 如果签名过期
		if expires < time.Now().Unix() && expires != 0 {
			logrus.WithField("Authorization", authorization).WithField("ttlUnix", authorizations[1]).Debugln("Authorization 头无效，签名已过期")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"data":    "",
				"message": "Authorization 头无效，签名已过期",
			})
			return
		}

		auth := &HMACAuth{
			CloudreveKey: []byte(pc.Conf.CloudreveKey),
		}

		// 获取待签名内容
		signContent := getSignContent(c.Request)

		// 生成签名
		generatedSign := auth.Sign(signContent, expires)

		// 检查签名是否匹配
		// 修复可能的前缀问题
		signatureTrimmed := signature
		// 如果签名中包含额外的前缀（如 "Cr "），尝试去除
		if parts := strings.Split(signature, " "); len(parts) > 1 {
			// 取最后一部分作为实际签名
			signatureTrimmed = parts[len(parts)-1]
		}

		if signatureTrimmed != generatedSign {
			logrus.WithFields(logrus.Fields{
				"Authorization":    authorization,
				"signature":        signature,
				"signatureTrimmed": signatureTrimmed,
				"generatedSign":    generatedSign,
				"signContent":      signContent,
			}).Debugln("Authorization 头无效，签名不匹配")

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"data":    "",
				"message": "Authorization 头无效，签名不匹配",
			})
			return
		}
	}
}
