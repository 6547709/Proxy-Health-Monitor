package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// =============================================================================
// 订阅地址获取与解析
// =============================================================================

// 排除的节点关键词（大写匹配）
var excludedPatterns = []string{
	"COMPATIBLE", "PASS", "应急节点", "应急续费", "DIRECT", "REJECT", "GLOBAL",
}

// fetchSubscription 从订阅地址获取内容
func fetchSubscription(url string, timeout int) ([]byte, error) {
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("获取订阅失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("订阅返回状态码: %d", resp.StatusCode)
	}

	// 限制读取大小为 10MB
	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, fmt.Errorf("读取订阅内容失败: %w", err)
	}

	return body, nil
}

// parseSubscription 解析订阅YAML，提取代理节点列表
// 使用 map 解析以保留完整配置，供 mihomo 创建代理适配器
func parseSubscription(data []byte) ([]ProxyNode, error) {
	var config struct {
		Proxies []map[string]any `yaml:"proxies"`
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("YAML解析失败: %w", err)
	}

	if len(config.Proxies) == 0 {
		return nil, fmt.Errorf("订阅内容中未找到代理节点")
	}

	// 过滤排除节点，提取关键字段
	var filtered []ProxyNode
	for _, raw := range config.Proxies {
		name, _ := raw["name"].(string)
		server, _ := raw["server"].(string)
		typ, _ := raw["type"].(string)

		// port 可能是 int 或 float64（YAML 解析）
		port := 0
		switch p := raw["port"].(type) {
		case int:
			port = p
		case float64:
			port = int(p)
		}

		if shouldExclude(name) {
			continue
		}
		if server == "" || port == 0 {
			continue
		}

		filtered = append(filtered, ProxyNode{
			Name:      name,
			Type:      typ,
			Server:    server,
			Port:      port,
			RawConfig: raw,
		})
	}

	return filtered, nil
}

// shouldExclude 检查节点名称是否应被排除
func shouldExclude(name string) bool {
	upperName := strings.ToUpper(name)
	for _, pattern := range excludedPatterns {
		if strings.Contains(upperName, pattern) {
			return true
		}
	}
	return false
}
