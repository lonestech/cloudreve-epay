package main

import (
	"crypto/md5"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// 过滤参数，生成签名时需删除 "sign" 和 "sign_type" 参数
func paramsFilter(params map[string]string) map[string]string {
	filtered := make(map[string]string)
	for k, v := range params {
		if k != "sign" && k != "sign_type" && v != "" {
			filtered[k] = v
		}
	}
	return filtered
}

// 对参数进行排序，返回排序后的 keys 和 values
func paramsSort(params map[string]string) ([]string, []string) {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	values := make([]string, len(keys))
	for i, k := range keys {
		values[i] = params[k]
	}

	return keys, values
}

// 生成待签名字符串, ["a", "b", "c"], ["d", "e", "f"] => "a=d&b=e&c=f"
func createUrlString(keys []string, values []string) string {
	urlString := ""
	for i, key := range keys {
		urlString += key + "=" + values[i] + "&"
	}
	// trim 掉最后的 &
	return strings.TrimSuffix(urlString, "&")
}

// 生成 加盐(商户 key) MD5 字符串
func md5String(urlString string, key string) string {
	digest := md5.Sum([]byte(urlString + "&key=" + key))
	return fmt.Sprintf("%x", digest)
}

// 生成易支付通知签名 - 与服务器使用相同的逻辑
func generateSign(params map[string]string, key string) string {
	fmt.Println("原始参数:")
	for k, v := range params {
		fmt.Printf("  %s = %s\n", k, v)
	}
	
	filtered := paramsFilter(params)
	fmt.Println("过滤后的参数:")
	for k, v := range filtered {
		fmt.Printf("  %s = %s\n", k, v)
	}
	
	keys, values := paramsSort(filtered)
	fmt.Println("排序后的键:")
	for i, k := range keys {
		fmt.Printf("  %d. %s\n", i+1, k)
	}
	
	urlString := createUrlString(keys, values)
	fmt.Println("生成的 URL 字符串:", urlString)
	fmt.Println("待签名字符串:", urlString + "&key=" + key)
	
	sign := md5String(urlString, key)
	fmt.Println("生成的签名:", sign)
	
	return sign
}

func main() {
	// 易支付密钥 - 与 .env 中的配置一致
	key := "SFDHSKHFJKDSHEUIFHU"
	
	// 使用已存在的订单号
	orderNo := "TEST_ORDER_1744189866"
	
	// 构建通知参数
	params := map[string]string{
		"pid":             "1010",
		"trade_no":        "10102025040915224738845",
		"out_trade_no":    orderNo,
		"type":            "alipay",
		"name":            "测试商品",
		"money":           "1.00",
		"trade_status":    "TRADE_SUCCESS",
		"inside_trade_no": "2025040922001429031405397942",
	}
	
	// 生成签名
	sign := generateSign(params, key)
	params["sign"] = sign
	params["sign_type"] = "MD5"
	
	fmt.Println("生成的签名:", sign)
	
	// 构建完整的通知 URL
	var builder strings.Builder
	builder.WriteString("http://localhost:4562/notify/")
	builder.WriteString(orderNo)
	builder.WriteString("?")
	
	queryParams := url.Values{}
	for k, v := range params {
		queryParams.Add(k, v)
	}
	
	builder.WriteString(queryParams.Encode())
	
	notifyURL := builder.String()
	fmt.Println("通知 URL:", notifyURL)
	fmt.Println("CURL 命令:")
	fmt.Printf("curl -X GET \"%s\"\n", notifyURL)
	
	// 输出创建订单的 CURL 命令
	fmt.Println("\n创建订单的 CURL 命令:")
	createOrderCmd := fmt.Sprintf("curl -X POST \"http://localhost:4562/cloudreve/purchase\" -H \"Content-Type: application/json\" -d '{\"name\":\"测试商品\", \"price\":\"1.00\", \"order_no\":\"%s\", \"amount\":100, \"notify_url\":\"http://localhost:4562/cloudreve/purchase\", \"return_url\":\"http://localhost:4562/return\"}'", orderNo)
	fmt.Println(createOrderCmd)
}
