package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"

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

// 过滤参数 - 与服务器使用相同的逻辑
func paramsFilter(params map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range params {
		if k != "sign" && k != "sign_type" && v != "" {
			result[k] = v
		}
	}
	return result
}

// 参数排序 - 与服务器使用相同的逻辑
func paramsSort(params map[string]string) ([]string, []string) {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	values := make([]string, 0, len(params))
	for _, k := range keys {
		values = append(values, params[k])
	}
	return keys, values
}

// 创建 URL 字符串 - 与服务器使用相同的逻辑
func createUrlString(keys, values []string) string {
	var buf strings.Builder
	for i := 0; i < len(keys); i++ {
		if i > 0 {
			buf.WriteString("&")
		}
		buf.WriteString(keys[i])
		buf.WriteString("=")
		buf.WriteString(values[i])
	}
	return buf.String()
}

// 生成 MD5 签名 - 与服务器使用相同的逻辑
func md5String(urlString, key string) string {
	h := md5.New()
	h.Write([]byte(urlString + "&key=" + key))
	return hex.EncodeToString(h.Sum(nil))
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
	orderNo := "TEST_ORDER_1744195816"
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

	// 使用与应用程序相同的密钥
	key := "SFDHSKHFJKDSHEUIFHU"

	// 构建通知参数
	params := map[string]string{
		"out_trade_no":   orderNo,
		"pid":            "1010",
		"trade_no":       "10102025040915224738845",
		"type":           "alipay",
		"name":           "测试商品",
		"money":          "1.00",
		"trade_status":   "TRADE_SUCCESS",
		"inside_trade_no": "2025040922001429031405397942",
	}

	// 生成签名
	sign := generateSign(params, key)
	params["sign"] = sign
	params["sign_type"] = "MD5"

	// 构建通知 URL
	notifyUrl := fmt.Sprintf("/notify/%s", orderNo)
	queryParams := make([]string, 0, len(params))
	for k, v := range params {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", k, v))
	}
	fullNotifyUrl := fmt.Sprintf("http://localhost:4562%s?%s", notifyUrl, strings.Join(queryParams, "&"))

	fmt.Println("\n通知 URL:", fullNotifyUrl)

	// 构建 CURL 命令
	curlCmd := fmt.Sprintf("curl -v -X GET \"%s\"", fullNotifyUrl)
	fmt.Println("\nCURL 命令:")
	fmt.Println(curlCmd)

	// 执行 CURL 命令
	fmt.Println("\n执行 CURL 命令:")
	cmd := exec.Command("bash", "-c", curlCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("执行 CURL 命令失败: %v\n", err)
	}
	fmt.Println(string(output))

	// 尝试直接发送 HTTP 请求
	fmt.Println("\n使用 Go HTTP 客户端发送请求:")
	resp, err := http.Get(fullNotifyUrl)
	if err != nil {
		fmt.Printf("发送 HTTP 请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// 读取响应
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	responseBody := buf.String()

	fmt.Printf("HTTP 状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", responseBody)
}
