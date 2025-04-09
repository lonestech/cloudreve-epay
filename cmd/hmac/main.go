package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

// RequestRawSign 待签名的HTTP请求
type RequestRawSign struct {
	Path   string
	Header string
	Body   string
}

// NewRequestSignString 返回JSON格式的待签名字符串
func NewRequestSignString(path, header, body string) string {
	req := RequestRawSign{
		Path:   path,
		Header: header,
		Body:   body,
	}
	res, _ := json.Marshal(req)
	return string(res)
}

func main() {
	// 使用与应用程序相同的密钥
	key := []byte("test234")
	
	// 要签名的路径
	path := "/cloudreve/purchase"
	
	// 请求正文
	body := `{"name":"测试商品", "price":"1.00", "order_no":"TEST_ORDER_` + strconv.FormatInt(time.Now().Unix(), 10) + `", "amount":100, "notify_url":"http://localhost:4562/cloudreve/purchase", "return_url":"http://localhost:4562/return"}`
	
	// 生成待签名内容（与服务器相同的方式）
	signContent := NewRequestSignString(path, "", body)
	
	// 生成过期时间（10分钟后）
	expires := time.Now().Add(10 * time.Minute).Unix()
	expireTimeStamp := strconv.FormatInt(expires, 10)
	
	// 生成签名
	h := hmac.New(sha256.New, key)
	_, err := io.WriteString(h, signContent+":"+expireTimeStamp)
	if err != nil {
		fmt.Println("Error generating signature:", err)
		os.Exit(1)
	}
	
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
	
	// 生成完整的 Authorization 头
	authHeader := fmt.Sprintf("Bearer %s:%s", signature, expireTimeStamp)
	
	fmt.Println("Authorization:", authHeader)
	fmt.Println("\nCURL 命令:")
	fmt.Printf("curl -X POST \"http://localhost:4562/cloudreve/purchase\" -H \"Content-Type: application/json\" -H \"Authorization: %s\" -d '%s'\n", authHeader, body)
}
