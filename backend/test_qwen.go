package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	// 从环境变量获取 API Key，或者直接在这里设置
	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	if apiKey == "" {
		// ���果环境变量没有，提示用户输入
		fmt.Print("请输入你的千问 API Key: ")
		fmt.Scanln(&apiKey)
	}

	if apiKey == "" {
		fmt.Println("错误：需要提供千问 API Key")
		return
	}

	// Coding Plan API 配置
	baseURL := "https://coding.dashscope.aliyuncs.com/v1"
	model := "qwen3.5-plus" // 使用配置文件中的模型

	// 构建请求
	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": "你好，请用一句话介绍你自己"},
		},
		"max_tokens":   100,
		"temperature":  0.7,
	}

	jsonBody, _ := json.Marshal(reqBody)

	// 发送请求
	req, err := http.NewRequest("POST", baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("\n状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应头: %v\n", resp.Header)
	fmt.Printf("响应体: %s\n", string(body))

	// 解析响应
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("\n解析响应失败: %v\n", err)
		return
	}

	if result.Error.Message != "" {
		fmt.Printf("\n❌ API 调用失败: %s (类型: %s)\n", result.Error.Message, result.Error.Type)
		return
	}

	if len(result.Choices) > 0 {
		fmt.Printf("\n✅ 千问 API 调用成功！\n")
		fmt.Printf("回复: %s\n", result.Choices[0].Message.Content)
	} else {
		fmt.Println("\n⚠️  API 返回成功但没有收到回复内容")
	}
}
