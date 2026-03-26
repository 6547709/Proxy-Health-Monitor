package main

import (
	"context"
	"fmt"
	"time"

	"github.com/metacubex/mihomo/adapter"
)

// =============================================================================
// 代理延迟测试 — 通过代理访问 generate_204 测量真实延迟
// =============================================================================

const (
	// testURL Clash/OpenClash 标准测试地址
	testURL = "http://www.gstatic.com/generate_204"
)

// testProxyLatency 使用 mihomo 的代理适配器，通过代理访问 generate_204 测量延迟
func testProxyLatency(rawConfig map[string]any, resolvedIP string, timeout time.Duration) (int, error) {
	// 使用已经解析好的 IP 替换域名，避免 mihomo 内部重复进行无缓存的 DNS 解析
	// 这通常是导致 CLI 测速比常驻运行的 Router 慢 100ms 左右的主要原因
	if resolvedIP != "" {
		// copy the map to avoid modifying the original
		newConfig := make(map[string]any)
		for k, v := range rawConfig {
			newConfig[k] = v
		}
		newConfig["server"] = resolvedIP
		rawConfig = newConfig
	}

	// 用 mihomo 从原始配置创建代理适配器
	proxy, err := adapter.ParseProxy(rawConfig)
	if err != nil {
		return 0, fmt.Errorf("创建代理适配器失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 使用 mihomo 原生的 URLTest 进行测速（与 OpenClash 方法完全一致：包括使用 HTTP HEAD 替代 GET，以及相同的计时起止点）
	// expectedStatus 传 nil 表示接受所有 2xx/3xx (204)
	delay, err := proxy.URLTest(ctx, testURL, nil)
	if err != nil {
		return 0, fmt.Errorf("代理延迟测试失败: %w", err)
	}

	return int(delay), nil
}
