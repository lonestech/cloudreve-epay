package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func main() {
	// 生成唯一的订单号
	orderNo := fmt.Sprintf("TEST_ORDER_%d", time.Now().Unix())
	
	// 构建 CURL 命令
	curlCmd := fmt.Sprintf(
		`curl -X POST "http://localhost:4562/cloudreve/purchase" -H "Content-Type: application/json" -d '{"name":"测试商品", "price":"1.00", "order_no":"%s", "amount":100, "notify_url":"http://example.com/callback", "return_url":"http://localhost:4562/return"}'`,
		orderNo,
	)
	
	// 打印 CURL 命令
	fmt.Println("创建订单的 CURL 命令:")
	fmt.Println(curlCmd)
	
	// 执行 CURL 命令
	cmd := exec.Command("bash", "-c", curlCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("执行 CURL 命令失败: %v\n", err)
		os.Exit(1)
	}
	
	// 打印输出
	fmt.Println("\n创建订单的响应:")
	fmt.Println(string(output))
	
	// 打印订单号
	fmt.Printf("\n创建的订单号: %s\n", orderNo)
	
	// 构建通知 CURL 命令
	notifyCurlCmd := fmt.Sprintf(
		`curl -X GET "http://localhost:4562/notify/%s?inside_trade_no=2025040922001429031405397942&money=1.00&name=%%E6%%B5%%8B%%E8%%AF%%95%%E5%%95%%86%%E5%%93%%81&out_trade_no=%s&pid=1010&sign_type=MD5&trade_no=10102025040915224738845&trade_status=TRADE_SUCCESS&type=alipay"`,
		orderNo, orderNo,
	)
	
	// 打印通知 CURL 命令
	fmt.Println("\n通知订单的 CURL 命令 (需要添加签名):")
	fmt.Println(notifyCurlCmd)
}
