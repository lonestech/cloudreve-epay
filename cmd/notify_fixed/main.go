package main

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

// 生成易支付通知签名
func generateSign(params map[string]string, key string) string {
	// 过滤参数
	filtered := make(map[string]string)
	for k, v := range params {
		if k != "sign" && k != "sign_type" && v != "" {
			filtered[k] = v
		}
	}

	// 按键名升序排序
	keys := make([]string, 0, len(filtered))
	for k := range filtered {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建待签名字符串
	var builder strings.Builder
	for _, k := range keys {
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(filtered[k])
		builder.WriteString("&")
	}
	
	// 去掉最后的 &
	signStr := strings.TrimSuffix(builder.String(), "&")
	fmt.Println("待签名字符串:", signStr + "&key=" + key)
	
	// 计算 MD5
	hash := md5.Sum([]byte(signStr + key))
	sign := fmt.Sprintf("%x", hash)
	
	return sign
}

// 生成 HMAC 签名
func generateHMACAuth(cloudreveKey []byte) string {
	// 使用当前时间戳作为过期时间（1分钟后过期）
	expireTime := time.Now().Add(time.Minute).Unix()
	expireTimeStamp := fmt.Sprintf("%d", expireTime)
	
	// 创建 HMAC
	h := hmac.New(sha256.New, cloudreveKey)
	h.Write([]byte("" + ":" + expireTimeStamp))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
	
	return "Bearer " + signature + ":" + expireTimeStamp
}

func main() {
	// 易支付密钥 - 与 .env.example 中的配置一致
	key := "SFDHSKHFJKDSHEUIFHU"
	
	// Cloudreve 密钥 - 与 .env.example 中的配置一致
	cloudreveKey := []byte("test234")
	
	// 订单号
	orderNo := "TEST_ORDER_1744185932"
	
	// 构建通知参数
	params := map[string]string{
		"inside_trade_no": "2025040922001429031405397942",
		"money":           "1.00",
		"name":            "测试商品",
		"out_trade_no":    orderNo,
		"pid":             "1010",
		"trade_no":        "10102025040915224738845",
		"trade_status":    "TRADE_SUCCESS",
		"type":            "alipay",
	}
	
	// 生成签名
	sign := generateSign(params, key)
	
	// 生成 HMAC 认证头
	authHeader := generateHMACAuth(cloudreveKey)
	fmt.Println("Authorization:", authHeader)
	
	// 构建完整的通知 URL
	var builder strings.Builder
	builder.WriteString("http://localhost:4560/notify/")
	builder.WriteString(orderNo)
	builder.WriteString("?")
	
	for k, v := range params {
		builder.WriteString(url.QueryEscape(k))
		builder.WriteString("=")
		builder.WriteString(url.QueryEscape(v))
		builder.WriteString("&")
	}
	
	builder.WriteString("sign=")
	builder.WriteString(sign)
	builder.WriteString("&sign_type=MD5")
	
	notifyURL := builder.String()
	fmt.Println("通知 URL:", notifyURL)
	fmt.Println("CURL 命令:")
	fmt.Printf("curl -X GET \"%s\"\n", notifyURL)
	
	// 打印带有 Authorization 头的 CURL 命令
	fmt.Println("\n带有 Authorization 头的 CURL 命令:")
	fmt.Printf("curl -X GET \"%s\" -H \"Authorization: %s\"\n", notifyURL, authHeader)
}
