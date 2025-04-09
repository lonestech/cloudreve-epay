package main

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/joho/godotenv"
	"github.com/topjohncian/cloudreve-pro-epay/internal/epay"
)

func main() {
	// 加载环境变量
	_ = godotenv.Load(".env.redis3")

	// 获取 EpayKey
	epayKey := os.Getenv("EPAY_KEY")
	if epayKey == "" {
		epayKey = "SFDHSKHFJKDSHEUIFHU" // 默认值，与通知工具中使用的相同
	}

	// 通知 URL
	notifyURL := "http://localhost:4562/notify/TEST_ORDER_1744189866?inside_trade_no=2025040922001429031405397942&money=1.00&name=%E6%B5%8B%E8%AF%95%E5%95%86%E5%93%81&out_trade_no=TEST_ORDER_1744189866&pid=1010&sign=4c0519b8f24900de82f90eaf2f2f4ad9&sign_type=MD5&trade_no=10102025040915224738845&trade_status=TRADE_SUCCESS&type=alipay"

	// 解析 URL
	parsedURL, err := url.Parse(notifyURL)
	if err != nil {
		fmt.Printf("解析 URL 失败: %v\n", err)
		os.Exit(1)
	}

	// 获取查询参数
	query := parsedURL.Query()
	params := make(map[string]string)
	for k, v := range query {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	// 获取签名
	providedSign := params["sign"]
	fmt.Printf("提供的签名: %s\n", providedSign)

	// 使用 epay 包中的方法生成签名
	serverSign := epay.GenerateSign(params, epayKey)
	fmt.Printf("服务器生成的签名: %s\n", serverSign)

	// 比较签名
	if serverSign == providedSign {
		fmt.Println("签名验证成功！")
	} else {
		fmt.Println("签名验证失败！")
		
		// 手动实现签名生成逻辑，以便与通知工具进行比较
		fmt.Println("\n手动实现签名生成逻辑:")
		
		// 过滤参数
		filtered := make(map[string]string)
		for k, v := range params {
			if k != "sign" && k != "sign_type" && v != "" {
				filtered[k] = v
			}
		}
		
		fmt.Println("过滤后的参数:")
		for k, v := range filtered {
			fmt.Printf("  %s = %s\n", k, v)
		}
		
		// 按键名升序排序
		keys := make([]string, 0, len(filtered))
		for k := range filtered {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		
		fmt.Println("排序后的键:")
		for i, k := range keys {
			fmt.Printf("  %d. %s\n", i+1, k)
		}
		
		// 构建待签名字符串
		var builder strings.Builder
		for i, k := range keys {
			if i > 0 {
				builder.WriteString("&")
			}
			builder.WriteString(k)
			builder.WriteString("=")
			builder.WriteString(filtered[k])
		}
		
		urlString := builder.String()
		fmt.Println("生成的 URL 字符串:", urlString)
		
		// 添加密钥
		signStr := urlString + "&key=" + epayKey
		fmt.Println("待签名字符串:", signStr)
		
		// 计算 MD5
		manualSign := epay.MD5String(urlString, epayKey)
		fmt.Println("手动生成的签名:", manualSign)
	}
}
