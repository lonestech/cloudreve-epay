package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

const (
	PaidOrderPrefix = "paid_order_"
)

func main() {
	// 加载环境变量
	_ = godotenv.Load(".env.redis3")

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

	// 要标记为已支付的订单号
	var orderNo string
	if len(os.Args) > 1 {
		orderNo = os.Args[1]
	} else {
		// 默认使用最近创建的订单号
		orderNo = "TEST_ORDER_1744199609"
	}

	// 标记订单为已支付
	paidKey := PaidOrderPrefix + orderNo
	err = rdb.Set(ctx, paidKey, true, 7*24*time.Hour).Err() // 保存 7 天
	if err != nil {
		fmt.Printf("无法标记订单为已支付: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("成功标记订单 %s 为已支付\n", orderNo)

	// 等待 1 秒后查询订单状态
	fmt.Println("\n等待 1 秒后查询订单状态...")
	time.Sleep(1 * time.Second)

	// 查询订单状态
	cmd := "go run cmd/query_order_status/main.go " + orderNo
	output, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		fmt.Printf("执行查询命令失败: %v\n", err)
		os.Exit(1)
	}

	// 打印查询结果
	fmt.Println("\n查询订单状态的结果:")
	fmt.Println(string(output))
}
