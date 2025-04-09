package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// 生成易支付的签名
func generateSign(params map[string]string, key string) string {
	var keys []string
	for k := range params {
		if k != "sign" && k != "sign_type" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var queryParts []string
	for _, k := range keys {
		queryParts = append(queryParts, k+"="+params[k])
	}
	queryString := strings.Join(queryParts, "&") + key

	hash := md5.Sum([]byte(queryString))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

func main() {
	// 获取订单号
	var orderNo string
	if len(os.Args) > 1 {
		orderNo = os.Args[1]
	} else {
		// 默认使用最近创建的订单号
		orderNo = "TEST_ORDER_1744199247"
	}

	// 构建回调参数
	params := map[string]string{
		"pid":             "1010",
		"trade_no":        "10102025040915224738845",
		"out_trade_no":    orderNo,
		"type":            "alipay",
		"name":            "测试商品",
		"money":           "1.00",
		"trade_status":    "TRADE_SUCCESS",
		"sign_type":       "MD5",
		"inside_trade_no": "2025040922001429031405397942",
	}

	// 使用与应用程序相同的密钥生成签名
	key := "SFDHSKHFJKDSHEUIFHU" // 从 .env.example 中获取的商家密钥
	sign := generateSign(params, key)
	params["sign"] = sign

	// 构建查询字符串
	var queryParts []string
	for k, v := range params {
		queryParts = append(queryParts, k+"="+v)
	}
	queryString := strings.Join(queryParts, "&")

	// 构建 CURL 命令
	curlCmd := fmt.Sprintf(`curl -X GET "http://localhost:4562/cloudreve/callback?%s"`, queryString)

	// 打印 CURL 命令
	fmt.Println("回调测试的 CURL 命令:")
	fmt.Println(curlCmd)

	// 执行 CURL 命令
	cmd := exec.Command("bash", "-c", curlCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("执行 CURL 命令失败: %v\n", err)
		os.Exit(1)
	}

	// 打印输出
	fmt.Println("\n回调测试的响应:")
	fmt.Println(string(output))

	// 等待 2 秒后查询订单状态
	fmt.Println("\n等待 2 秒后查询订单状态...")
	time.Sleep(2 * time.Second)

	// 查询订单状态
	queryCmd := exec.Command("go", "run", "cmd/query_order_status/main.go", orderNo)
	queryOutput, err := queryCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("执行查询命令失败: %v\n", err)
		os.Exit(1)
	}

	// 打印查询结果
	fmt.Println("\n查询订单状态的结果:")
	fmt.Println(string(queryOutput))
}
