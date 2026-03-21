package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// =============================================================================
// 版本信息
// =============================================================================

const Version = "2.1.0"

// =============================================================================
// JSON 输出结构（供程序调用）
// =============================================================================

// JSONOutput JSON 格式的检测结果
type JSONOutputData struct {
	Version    string           `json:"version"`
	UpdateTime string           `json:"update_time"`
	Stats      NodeStats        `json:"stats"`
	Regions    []JSONRegionData `json:"regions"`
}

type JSONRegionData struct {
	Name   string           `json:"name"`
	Flag   string           `json:"flag"`
	NameEN string           `json:"name_en"`
	Stats  NodeStats        `json:"stats"`
	Nodes  []JSONNodeData   `json:"nodes"`
}

type JSONNodeData struct {
	Name         string   `json:"name"`
	Server       string   `json:"server"`
	Port         int      `json:"port"`
	DNSResolved  bool     `json:"dns_resolved"`
	ResolvedIPs  []string `json:"resolved_ips,omitempty"`
	TCPConnected bool     `json:"tcp_connected"`
	Latency      int      `json:"latency"`
	Category     string   `json:"category"`
	Error        string   `json:"error,omitempty"`
}

// printJSON 以 JSON 格式输出检测结果
func printJSON(regions []RegionData, stats NodeStats) {
	output := JSONOutputData{
		Version:    Version,
		UpdateTime: time.Now().Format("2006-01-02T15:04:05Z07:00"),
		Stats:      stats,
	}
	for _, r := range regions {
		jr := JSONRegionData{
			Name:   r.Name,
			Flag:   r.Flag,
			NameEN: r.NameEN,
			Stats:  r.Stats,
		}
		for _, n := range r.Results {
			jr.Nodes = append(jr.Nodes, JSONNodeData{
				Name:         cleanNodeName(n.Node.Name),
				Server:       n.Node.Server,
				Port:         n.Node.Port,
				DNSResolved:  n.DNSResolved,
				ResolvedIPs:  n.ResolvedIPs,
				TCPConnected: n.TCPConnected,
				Latency:      n.Latency,
				Category:     n.Category,
				Error:        n.Error,
			})
		}
		output.Regions = append(output.Regions, jr)
	}
	data, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(data))
}

// =============================================================================
// 主入口
// =============================================================================

func main() {
	config := parseConfig()

	// 验证必填参数
	if config.SubscriptionURL == "" {
		fmt.Println()
		fmt.Printf("  %s%s错误: 订阅地址不能为空%s\n\n", colorBold, colorRed, colorReset)
		fmt.Println("  使用方式:")
		fmt.Printf("    %sproxy-monitor -url \"https://sub.example.com/link/xxx?clash=2\"%s\n", colorCyan, colorReset)
		fmt.Println()
		fmt.Println("  或通过环境变量:")
		fmt.Printf("    %sexport SUBSCRIBE_URL=\"https://sub.example.com/link/xxx?clash=2\"%s\n", colorCyan, colorReset)
		fmt.Printf("    %sproxy-monitor%s\n", colorCyan, colorReset)
		fmt.Println()
		fmt.Println("  更多选项:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("  环境变量说明:")
		fmt.Println("    SUBSCRIBE_URL     订阅地址")
		fmt.Println("    INTERVAL          刷新间隔秒数（默认 60）")
		fmt.Println("    CONCURRENT        最大并发数（默认 20）")
		fmt.Println("    TIMEOUT           超时秒数（默认 10）")
		fmt.Println("    NO_COLOR          禁用颜色（设为 1 或 true）")
		fmt.Println("    JSON_OUTPUT       JSON格式输出（设为 1 或 true）")
		fmt.Println()
		fmt.Println("  注：CLI 参数优先级高于环境变量")
		os.Exit(1)
	}

	if config.Watch {
		// 持续监控模式
		runWatch(config)
	} else {
		// 默认：单次执行
		runOnce(config)
	}
}

// parseConfig 解析配置：CLI 参数优先，环境变量兜底
func parseConfig() AppConfig {
	var config AppConfig

	// 定义 CLI 参数
	url := flag.String("url", "", "订阅地址 (env: SUBSCRIBE_URL)")
	interval := flag.Int("interval", 0, "刷新间隔秒数，默认 60 (env: INTERVAL)")
	concurrent := flag.Int("concurrent", 0, "最大并发检测数，默认 20 (env: CONCURRENT)")
	timeout := flag.Int("timeout", 0, "单节点超时秒数，默认 10 (env: TIMEOUT)")
	watch := flag.Bool("watch", false, "持续监控模式（默认执行一次）")
	noColor := flag.Bool("no-color", false, "禁用颜色输出 (env: NO_COLOR)")
	jsonOutput := flag.Bool("json", false, "JSON格式输出，适合程序调用 (env: JSON_OUTPUT)")
	showVersion := flag.Bool("version", false, "显示版本号")

	flag.Parse()

	if *showVersion {
		fmt.Printf("proxy-monitor v%s\n", Version)
		os.Exit(0)
	}

	// 订阅地址：CLI > 环境变量
	config.SubscriptionURL = getStringConfig(*url, "SUBSCRIBE_URL", "")

	// 刷新间隔
	config.Interval = getIntConfig(*interval, "INTERVAL", 60)

	// 最大并发
	config.Concurrent = getIntConfig(*concurrent, "CONCURRENT", 20)

	// 超时
	config.Timeout = getIntConfig(*timeout, "TIMEOUT", 10)

	// 持续监控模式
	config.Watch = *watch

	// 颜色
	config.NoColor = getBoolConfig(*noColor, "NO_COLOR", false)

	// JSON 输出
	config.JSONOutput = getBoolConfig(*jsonOutput, "JSON_OUTPUT", false)

	return config
}

// getStringConfig CLI参数 > 环境变量 > 默认值
func getStringConfig(cliValue, envKey, defaultVal string) string {
	if cliValue != "" {
		return cliValue
	}
	if envVal := os.Getenv(envKey); envVal != "" {
		return envVal
	}
	return defaultVal
}

// getIntConfig CLI参数 > 环境变量 > 默认值
func getIntConfig(cliValue int, envKey string, defaultVal int) int {
	if cliValue > 0 {
		return cliValue
	}
	if envVal := os.Getenv(envKey); envVal != "" {
		if v, err := strconv.Atoi(envVal); err == nil && v > 0 {
			return v
		}
	}
	return defaultVal
}

// getBoolConfig CLI参数 > 环境变量 > 默认值
func getBoolConfig(cliValue bool, envKey string, defaultVal bool) bool {
	if cliValue {
		return true
	}
	if envVal := os.Getenv(envKey); envVal != "" {
		lower := strings.ToLower(envVal)
		return lower == "1" || lower == "true" || lower == "yes"
	}
	return defaultVal
}

// runOnce 执行一次检测并退出
func runOnce(config AppConfig) {
	regions, stats, err := doCheck(config)
	if err != nil {
		if config.JSONOutput {
			fmt.Printf(`{"error":"%s"}\n`, err.Error())
		} else {
			renderError(err.Error())
		}
		os.Exit(1)
	}

	if config.JSONOutput {
		printJSON(regions, stats)
		return
	}

	updateTime := time.Now().Format("15:04:05")
	renderDashboard(regions, stats, config.SubscriptionURL, Version, updateTime, 0, false)
}

// runWatch 持续监控模式
func runWatch(config AppConfig) {
	// 捕获退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		regions, stats, err := doCheck(config)
		if err != nil {
			clearScreen()
			renderError(err.Error())
			fmt.Printf("  %s%d 秒后重试...%s\n", colorGray, config.Interval, colorReset)
		} else {
			updateTime := time.Now().Format("15:04:05")
			renderDashboard(regions, stats, config.SubscriptionURL, Version, updateTime, config.Interval, true)
		}

		// 等待刷新间隔或退出信号
		select {
		case <-sigCh:
			fmt.Printf("\n  %s👋 再见！%s\n\n", colorCyan, colorReset)
			return
		case <-time.After(time.Duration(config.Interval) * time.Second):
			// 继续下一轮
		}
	}
}

// doCheck 执行完整的检测流程
func doCheck(config AppConfig) ([]RegionData, NodeStats, error) {
	// 第一步：获取订阅
	fmt.Printf("\n  %s📡 正在获取订阅...%s", colorCyan, colorReset)
	data, err := fetchSubscription(config.SubscriptionURL, config.Timeout*2)
	if err != nil {
		return nil, NodeStats{}, err
	}

	// 第二步：解析节点
	nodes, err := parseSubscription(data)
	if err != nil {
		return nil, NodeStats{}, err
	}
	fmt.Printf("\r  %s📡 获取到 %d 个节点，开始检测...%s\n", colorCyan, len(nodes), colorReset)

	// 第三步：并发检测
	timeout := time.Duration(config.Timeout) * time.Second
	results := checkAllNodes(nodes, config.Concurrent, timeout, renderProgress)

	// 第四步：区域分组
	regions := groupByRegion(results)
	stats := calcTotalStats(regions)

	return regions, stats, nil
}
