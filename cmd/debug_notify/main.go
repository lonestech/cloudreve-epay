package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net/url"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
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
	orderNo := "TEST_ORDER_1744189866"
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
	
	// 模拟金额验证
	amount := decimal.NewFromInt(int64(retrievedOrder.Amount)).Div(decimal.NewFromInt(100))
	realAmount, err := decimal.NewFromString("1.00")
	if err != nil {
		fmt.Printf("无法解析订单金额: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("订单金额: %d\n", retrievedOrder.Amount)
	fmt.Printf("计算后的金额: %s\n", amount.String())
	fmt.Printf("通知中的金额: %s\n", realAmount.String())
	
	if !realAmount.Equal(amount) {
		fmt.Println("订单金额不符")
	} else {
		fmt.Println("订单金额匹配")
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
	
	// 打印通知 URL 和 CURL 命令
	fmt.Println("\n通知 URL:", notifyURL)
	fmt.Println("\nCURL 命令:")
	fmt.Printf("curl -X GET \"%s\"\n", notifyURL)
}
