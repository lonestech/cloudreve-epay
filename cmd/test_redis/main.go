package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

// PurchaseRequest 模拟订单请求结构
type PurchaseRequest struct {
	Name      string `json:"name"`
	OrderNo   string `json:"order_no"`
	NotifyUrl string `json:"notify_url"`
	Amount    int    `json:"amount"`
}

const (
	PurchaseSessionPrefix = "purchase_session_"
)

func main() {
	// 加载环境变量
	_ = godotenv.Load(".env")

	// 连接 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis 地址
		Password: "",               // 无密码
		DB:       0,                // 默认 DB
	})

	ctx := context.Background()

	// 测试 Redis 连接
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("无法连接到 Redis: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("成功连接到 Redis")

	// 创建测试订单
	orderNo := "TEST_ORDER_1744186855"
	testOrder := &PurchaseRequest{
		Name:      "测试商品",
		OrderNo:   orderNo,
		NotifyUrl: "http://localhost:4562/cloudreve/purchase",
		Amount:    100,
	}

	// 将订单转换为 JSON
	orderJSON, err := json.Marshal(testOrder)
	if err != nil {
		fmt.Printf("无法序列化订单: %v\n", err)
		os.Exit(1)
	}

	// 存储到 Redis
	cacheKey := PurchaseSessionPrefix + orderNo
	err = rdb.Set(ctx, cacheKey, orderJSON, 24*time.Hour).Err()
	if err != nil {
		fmt.Printf("无法存储订单到 Redis: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("已将订单 %s 存储到 Redis\n", orderNo)

	// 从 Redis 中获取订单
	orderData, err := rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		fmt.Printf("无法从 Redis 获取订单: %v\n", err)
		os.Exit(1)
	}

	// 解析订单数据
	var retrievedOrder PurchaseRequest
	err = json.Unmarshal([]byte(orderData), &retrievedOrder)
	if err != nil {
		fmt.Printf("无法解析订单数据: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("成功从 Redis 获取订单 %s\n", orderNo)
	fmt.Printf("订单信息: %+v\n", retrievedOrder)

	// 生成通知 URL 和 CURL 命令
	fmt.Println("\n通知 URL 和 CURL 命令:")
	notifyURL := fmt.Sprintf("http://localhost:4562/notify/%s?money=1.00&name=%%E6%%B5%%8B%%E8%%AF%%95%%E5%%95%%86%%E5%%93%%81&pid=1010&trade_no=10102025040915224738845&trade_status=TRADE_SUCCESS&type=alipay&out_trade_no=%s&inside_trade_no=2025040922001429031405397942&sign=ea3d32b87314f5cbb0b83ae438ee14aa&sign_type=MD5", orderNo, orderNo)
	fmt.Printf("通知 URL: %s\n", notifyURL)
	fmt.Printf("CURL 命令: curl -X GET \"%s\"\n", notifyURL)
}
