package main

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// =============================================================================
// 健康检测：DNS 解析 → TCP 连通 → 代理延迟（generate_204）
// =============================================================================

// checkNode 对单个节点执行三步检测
func checkNode(node ProxyNode, timeout time.Duration) CheckResult {
	result := CheckResult{
		Node:     node,
		Category: "fault",
	}

	// 第一步：DNS 解析
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resolver := &net.Resolver{}
	ips, err := resolver.LookupHost(ctx, node.Server)
	if err != nil {
		result.Error = fmt.Sprintf("DNS解析失败: %v", err)
		return result
	}
	result.DNSResolved = true
	result.ResolvedIPs = ips

	// 第二步：TCP 连通检测
	address := net.JoinHostPort(node.Server, fmt.Sprintf("%d", node.Port))
	startTcp := time.Now()
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		result.Error = fmt.Sprintf("TCP连接失败: %v", err)
		return result
	}
	result.TCPPing = int(time.Since(startTcp).Milliseconds())
	conn.Close()
	result.TCPConnected = true

	// 第三步：代理延迟测试（通过代理访问 generate_204）
	// 将解析好的 IP 传入，避免 mihomo 内置重新解析 DNS 导致的额外延迟
	latency, err := testProxyLatency(node.RawConfig, result.ResolvedIPs[0], timeout)
	if err != nil {
		result.Error = fmt.Sprintf("代理测试失败: %v", err)
		// TCP 是通的但代理不工作，标记为故障
		result.Category = "fault"
		return result
	}

	result.Latency = latency
	result.Category = categorizeDelay(latency)

	return result
}

// checkAllNodes 并发检测所有节点
func checkAllNodes(nodes []ProxyNode, concurrent int, timeout time.Duration, progressFn func(done, total int)) []CheckResult {
	results := make([]CheckResult, len(nodes))
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrent) // 并发控制信号量

	var mu sync.Mutex
	done := 0

	for i, node := range nodes {
		wg.Add(1)
		go func(idx int, n ProxyNode) {
			defer wg.Done()
			sem <- struct{}{}        // 获取信号量
			defer func() { <-sem }() // 释放信号量

			results[idx] = checkNode(n, timeout)

			mu.Lock()
			done++
			if progressFn != nil {
				progressFn(done, len(nodes))
			}
			mu.Unlock()
		}(i, node)
	}

	wg.Wait()
	return results
}

// categorizeDelay 根据延迟分类
func categorizeDelay(delay int) string {
	if delay <= 0 {
		return "fault"
	} else if delay < 150 {
		return "fast"
	} else if delay < 220 {
		return "normal"
	} else if delay < 310 {
		return "high_latency"
	}
	return "fault"
}
