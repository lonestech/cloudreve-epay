package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
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
	
	// 获取订单号
	var orderNo string
	if len(os.Args) > 1 {
		orderNo = os.Args[1]
	} else {
		// 默认使用最近创建的订单号
		orderNo = "TEST_ORDER_1744199247"
	}
	
	// 要签名的路径（包含查询参数）
	path := "/cloudreve/purchase"
	
	// 生成待签名内容（与服务器相同的方式）
	signContent := NewRequestSignString(path, "", "")
	
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
	
	// 构建 CURL 命令
	curlCmd := fmt.Sprintf(`curl -X GET "http://localhost:4562/cloudreve/purchase?order_no=%s" -H "Authorization: %s"`, orderNo, authHeader)
	
	// 打印 CURL 命令
	fmt.Println("查询订单状态的 CURL 命令:")
	fmt.Println(curlCmd)
	
	// 执行 CURL 命令
	cmd := exec.Command("bash", "-c", curlCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("执行 CURL 命令失败: %v\n", err)
		os.Exit(1)
	}
	
	// 打印输出
	fmt.Println("\n查询订单状态的响应:")
	fmt.Println(string(output))
}
