package main

import (
	"crypto/md5"
	"fmt"
	"net/url"
	"sort"
	"strings"
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

func main() {
	// 易支付密钥 - 与 .env.example 中的配置一致
	key := "SFDHSKHFJKDSHEUIFHU"
	
	// 订单号
	orderNo := "TEST_ORDER_1744185616"
	
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
	
	// 构建完整的通知 URL
	var builder strings.Builder
	builder.WriteString("http://localhost:4560/cloudreve/purchase?")
	
	for k, v := range params {
		builder.WriteString(url.QueryEscape(k))
		builder.WriteString("=")
		builder.WriteString(url.QueryEscape(v))
		builder.WriteString("&")
	}
	
	builder.WriteString("sign=")
	builder.WriteString(sign)
	builder.WriteString("&sign_type=MD5")
	
	fmt.Println("通知 URL:", builder.String())
	fmt.Println("CURL 命令:")
	fmt.Printf("curl -X GET \"%s\"\n", builder.String())
}
