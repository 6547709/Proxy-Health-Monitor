package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/metacubex/mihomo/adapter"
	C "github.com/metacubex/mihomo/constant"
)

// =============================================================================
// 代理延迟测试 — 通过代理访问 generate_204 测量真实延迟
// =============================================================================

const (
	// testURL Clash/OpenClash 标准测试地址
	testURL = "http://www.gstatic.com/generate_204"
)

// testProxyLatency 使用 mihomo 的代理适配器，通过代理访问 generate_204 测量延迟
func testProxyLatency(rawConfig map[string]any, timeout time.Duration) (int, error) {
	// 用 mihomo 从原始配置创建代理适配器
	proxy, err := adapter.ParseProxy(rawConfig)
	if err != nil {
		return 0, fmt.Errorf("创建代理适配器失败: %w", err)
	}

	// 创建通过代理的 HTTP Transport
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, portStr, _ := net.SplitHostPort(addr)
			port, _ := strconv.Atoi(portStr)
			metadata := &C.Metadata{
				NetWork: C.TCP,
				Host:    host,
				DstPort: uint16(port),
			}
			return proxy.DialContext(ctx, metadata)
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
		// 不跟随重定向
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// 测量通过代理访问 generate_204 的延迟
	start := time.Now()
	resp, err := client.Get(testURL)
	elapsed := time.Since(start)

	if err != nil {
		return 0, fmt.Errorf("代理延迟测试失败: %w", err)
	}
	defer resp.Body.Close()

	// 204 No Content 或 200 OK 都算成功
	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		return 0, fmt.Errorf("非预期状态码: %d", resp.StatusCode)
	}

	return int(elapsed.Milliseconds()), nil
}
