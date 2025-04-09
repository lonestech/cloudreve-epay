package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/topjohncian/cloudreve-pro-epay/internal/controller"
)

const (
	PurchaseSessionPrefix = "purchase_session_"
)

// item 对应 Redis 缓存中的存储结构
type item struct {
	Value interface{}
}

// 反序列化函数
func deserializer(value []byte) (interface{}, error) {
	var res item
	buffer := bytes.NewReader(value)
	dec := gob.NewDecoder(buffer)
	err := dec.Decode(&res)
	if err != nil {
		return nil, err
	}
	return res.Value, nil
}

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

	// 要查询的订单号
	orderNo := "TEST_ORDER_1744199609"
	cacheKey := PurchaseSessionPrefix + orderNo

	// 从 Redis 中获取订单
	orderDataBytes, err := rdb.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			fmt.Printf("订单 %s 不存在于 Redis 中\n", orderNo)
		} else {
			fmt.Printf("无法从 Redis 获取订单: %v\n", err)
		}
		os.Exit(1)
	}

	// 使用 gob 反序列化
	decodedValue, err := deserializer(orderDataBytes)
	if err != nil {
		fmt.Printf("无法反序列化订单数据: %v\n", err)
		os.Exit(1)
	}
	
	// 尝试转换为 PurchaseRequest 类型
	retrievedOrder, ok := decodedValue.(*controller.PurchaseRequest)
	if !ok {
		fmt.Printf("无法转换为 PurchaseRequest 类型: %T\n", decodedValue)
		os.Exit(1)
	}

	fmt.Printf("成功从 Redis 获取订单 %s\n", orderNo)
	fmt.Printf("订单信息: %+v\n", retrievedOrder)

	// 检查订单是否已支付
	paidKey := "paid_order_" + orderNo
	exists, err := rdb.Exists(ctx, paidKey).Result()
	if err != nil {
		fmt.Printf("无法检查订单支付状态: %v\n", err)
		os.Exit(1)
	}

	if exists == 1 {
		fmt.Println("订单已标记为已支付")
	} else {
		fmt.Println("订单未标记为已支付")	
	}
}
