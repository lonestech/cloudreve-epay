package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/topjohncian/cloudreve-pro-epay/internal/appconf"
	"github.com/topjohncian/cloudreve-pro-epay/internal/cache"
	"github.com/topjohncian/cloudreve-pro-epay/internal/controller"
)

func main() {
	// 加载环境变量
	_ = godotenv.Load(".env")

	// 解析配置
	conf, _ := appconf.Parse()

	// 创建缓存
	cacheStore := cache.NewMemoStore()

	// 检查订单号
	orderNo := "TEST_ORDER_1744185932"
	cacheKey := controller.PurchaseSessionPrefix + orderNo

	// 创建一个测试订单并存入缓存
	testOrder := &controller.PurchaseRequest{
		Name:      "测试商品",
		OrderNo:   orderNo,
		NotifyUrl: "http://localhost:4560/cloudreve/purchase",
		Amount:    100,
	}

	// 存入缓存
	err := cacheStore.Set(cacheKey, testOrder, 3600)
	if err != nil {
		fmt.Printf("存入缓存失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("已将订单 %s 存入缓存\n", orderNo)

	// 从缓存中获取
	value, ok := cacheStore.Get(cacheKey)
	if !ok {
		fmt.Printf("从缓存中获取订单 %s 失败\n", orderNo)
		os.Exit(1)
	}

	// 尝试转换为 PurchaseRequest
	order, ok := value.(*controller.PurchaseRequest)
	if !ok {
		fmt.Printf("订单信息类型转换失败: %T\n", value)
		os.Exit(1)
	}

	fmt.Printf("成功从缓存中获取订单 %s\n", orderNo)
	fmt.Printf("订单信息: %+v\n", order)

	// 直接查看缓存的内部结构
	fmt.Println("\n检查缓存的内部结构:")
	memoStore, ok := cacheStore.(*cache.MemoStore)
	if !ok {
		fmt.Println("无法转换为 MemoStore 类型")
		return
	}
	memoStore.Store.Range(func(key, value interface{}) bool {
		fmt.Printf("键: %v, 值类型: %T\n", key, value)
		return true
	})
}
