package main

import (
	"fmt"
	"strings"
)

// =============================================================================
// 终端渲染 — ANSI 彩色输出
// =============================================================================

// ANSI 颜色代码
const (
	colorReset   = "\033[0m"
	colorBold    = "\033[1m"
	colorDim     = "\033[2m"
	colorCyan    = "\033[36m"
	colorMagenta = "\033[35m"
	colorBlue    = "\033[38;5;39m"  // 高速 #00a8ff
	colorGreen   = "\033[38;5;48m"  // 正常 #00ff88
	colorYellow  = "\033[38;5;220m" // 高延迟 #ffcc00
	colorRed     = "\033[38;5;204m" // 故障 #ff3366
	colorWhite   = "\033[37m"
	colorGray    = "\033[90m"
)

// clearScreen 清屏
func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

// colorByCategory 根据分类返回对应颜色
func colorByCategory(category string) string {
	switch category {
	case "fast":
		return colorBlue
	case "normal":
		return colorGreen
	case "high_latency":
		return colorYellow
	case "fault":
		return colorRed
	default:
		return colorGray
	}
}

// renderHeader 渲染标题栏
func renderHeader(subURL string, version string) {
	width := 56
	border := strings.Repeat("═", width-2)
	fmt.Printf("%s%s╔%s╗%s\n", colorBold, colorCyan, border, colorReset)
	title := "🌐 全球代理节点健康监控"
	fmt.Printf("%s%s║%s  %s v%s%s", colorBold, colorCyan, colorReset, title, version, colorReset)
	// 填充到右边框
	titleLen := 34 + len(version) // 估算显示宽度
	padding := width - 2 - titleLen
	if padding > 0 {
		fmt.Print(strings.Repeat(" ", padding))
	}
	fmt.Printf("%s%s║%s\n", colorBold, colorCyan, colorReset)

	// 订阅地址行（截断显示）
	displayURL := subURL
	maxURLLen := width - 8
	if len(displayURL) > maxURLLen {
		displayURL = displayURL[:maxURLLen-3] + "..."
	}
	fmt.Printf("%s%s║%s  %s📡 %s%s", colorBold, colorCyan, colorReset, colorGray, displayURL, colorReset)
	urlLineLen := 5 + len(displayURL)
	urlPad := width - 2 - urlLineLen
	if urlPad > 0 {
		fmt.Print(strings.Repeat(" ", urlPad))
	}
	fmt.Printf("%s%s║%s\n", colorBold, colorCyan, colorReset)

	fmt.Printf("%s%s╚%s╝%s\n", colorBold, colorCyan, border, colorReset)
}

// renderStats 渲染统计面板
func renderStats(stats NodeStats) {
	fmt.Println()
	fmt.Printf("  %s🚀 高速: %-4d%s   %s✅ 正常: %-4d%s   %s⚠️  高延迟: %-4d%s   %s❌ 故障: %-4d%s\n",
		colorBlue, stats.Fast, colorReset,
		colorGreen, stats.Normal, colorReset,
		colorYellow, stats.HighLatency, colorReset,
		colorRed, stats.Fault, colorReset,
	)

	healthy := stats.Fast + stats.Normal
	pct := 0
	if stats.Total > 0 {
		pct = healthy * 100 / stats.Total
	}
	fmt.Printf("  %s健康度: 高速+正常 = %d 节点 (%d%%)%s\n",
		colorGreen, healthy, pct, colorReset)
	fmt.Println()
}

// renderRegionSummary 渲染区域汇总表（在详细列表之前）
func renderRegionSummary(regions []RegionData) {
	sep := strings.Repeat("━", 56)
	fmt.Printf("%s%s%s\n", colorCyan, sep, colorReset)
	fmt.Printf("  %s%s各区域概览%s\n", colorBold, colorCyan, colorReset)
	thinSep := strings.Repeat("─", 56)
	fmt.Printf("%s%s%s\n", colorGray, thinSep, colorReset)

	for _, r := range regions {
		flag := padToWidth(r.Flag+" "+r.Name, 14)
		healthy := r.Stats.Fast + r.Stats.Normal
		pct := 0
		if r.Stats.Total > 0 {
			pct = healthy * 100 / r.Stats.Total
		}
		// 根据健康度选颜色
		pctColor := colorRed
		if pct >= 80 {
			pctColor = colorGreen
		} else if pct >= 50 {
			pctColor = colorYellow
		}
		fmt.Printf("  %s  %s🚀%-3d %s✅%-3d %s⚠️ %-3d %s❌%-3d  %s%3d%%%s  %s(%d节点)%s\n",
			flag,
			colorBlue, r.Stats.Fast,
			colorGreen, r.Stats.Normal,
			colorYellow, r.Stats.HighLatency,
			colorRed, r.Stats.Fault,
			pctColor, pct, colorReset,
			colorGray, r.Stats.Total, colorReset,
		)
	}
	fmt.Println()
}

// renderRegion 渲染单个区域
func renderRegion(region RegionData) {
	sep := strings.Repeat("━", 56)
	fmt.Printf("%s%s%s\n", colorCyan, sep, colorReset)

	// 区域标题
	fmt.Printf("%s%s %s%s %s%s%s%s %d 节点%s\n",
		colorBold, region.Flag,
		colorWhite, region.Name,
		colorGray, region.NameEN, colorReset,
		colorCyan, region.Stats.Total, colorReset,
	)

	// 区域统计
	fmt.Printf("  %s🚀 %d%s  %s✅ %d%s  %s⚠️  %d%s  %s❌ %d%s\n",
		colorBlue, region.Stats.Fast, colorReset,
		colorGreen, region.Stats.Normal, colorReset,
		colorYellow, region.Stats.HighLatency, colorReset,
		colorRed, region.Stats.Fault, colorReset,
	)

	// 分隔线
	thinSep := strings.Repeat("─", 56)
	fmt.Printf("%s%s%s\n", colorGray, thinSep, colorReset)

	// 节点列表
	for _, r := range region.Results {
		renderNode(r)
	}
}

// truncateDisplay 按显示宽度截断字符串（中文/emoji占2字符宽度）
func truncateDisplay(s string, maxWidth int) string {
	runes := []rune(s)
	width := 0
	for i, r := range runes {
		w := 1
		// 中文、日文、韩文等宽字符占2个显示宽度
		if r > 0x2E80 {
			w = 2
		}
		if width+w > maxWidth-3 {
			return string(runes[:i]) + "..."
		}
		width += w
	}
	return s
}

// padToWidth 将字符串填充到指定显示宽度
func padToWidth(s string, targetWidth int) string {
	runes := []rune(s)
	width := 0
	for _, r := range runes {
		if r > 0x2E80 {
			width += 2
		} else {
			width++
		}
	}
	if width < targetWidth {
		return s + strings.Repeat(" ", targetWidth-width)
	}
	return s
}

// renderNode 渲染单个节点
func renderNode(r CheckResult) {
	name := cleanNodeName(r.Node.Name)
	name = truncateDisplay(name, 22)
	name = padToWidth(name, 22)

	// DNS 状态
	dnsStatus := fmt.Sprintf("%s✓%s", colorGreen, colorReset)
	if !r.DNSResolved {
		dnsStatus = fmt.Sprintf("%s✗%s", colorRed, colorReset)
	}

	// TCP 状态与延迟
	tcpStatus := fmt.Sprintf("%s✗%s", colorRed, colorReset)
	if r.TCPConnected {
		tcpStatus = fmt.Sprintf("%sTCP %3dms%s", colorCyan, r.TCPPing, colorReset)
	}

	// 延迟（HTTP）
	latencyStr := "HTTP   ---"
	if r.TCPConnected && r.Latency > 0 {
		latencyStr = fmt.Sprintf("HTTP %4dms", r.Latency)
	}

	// 分类图标
	catIcon := "❌"
	switch r.Category {
	case "fast":
		catIcon = "🚀"
	case "normal":
		catIcon = "✅"
	case "high_latency":
		catIcon = "⚠️"
	}

	clr := colorByCategory(r.Category)

	// 解析IP（截短显示）
	ipStr := ""
	if r.DNSResolved && len(r.ResolvedIPs) > 0 {
		ip := r.ResolvedIPs[0]
		if len(ip) > 15 {
			ip = ip[:12] + "..."
		}
		ipStr = fmt.Sprintf(" %s[%s]%s", colorGray, ip, colorReset)
	}

	fmt.Printf("  %s  DNS %s  %s  %s%s%s  %s%s\n",
		name, dnsStatus, tcpStatus, clr, latencyStr, colorReset, catIcon, ipStr)
}

// renderProgress 渲染检测进度
func renderProgress(done, total int) {
	pct := 0
	if total > 0 {
		pct = done * 100 / total
	}
	barWidth := 30
	filled := barWidth * done / total
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	fmt.Printf("\r  %s检测进度: [%s%s%s] %d/%d (%d%%)%s",
		colorCyan, colorBlue, bar, colorCyan, done, total, pct, colorReset)
	if done >= total {
		fmt.Println() // 完成后换行
	}
}

// renderDashboard 渲染完整 Dashboard
func renderDashboard(regions []RegionData, stats NodeStats, subURL, version, updateTime string, interval int, watchMode bool) {
	clearScreen()
	renderHeader(subURL, version)
	renderStats(stats)

	// 区域汇总（在详细列表之前）
	renderRegionSummary(regions)

	for _, region := range regions {
		renderRegion(region)
	}

	fmt.Println()
	bottomSep := strings.Repeat("━", 56)
	fmt.Printf("%s%s%s\n", colorCyan, bottomSep, colorReset)
	fmt.Printf("  %s更新时间: %s%s\n", colorGray, updateTime, colorReset)
	if watchMode {
		fmt.Printf("  %s下次刷新: %d秒后  |  Ctrl+C 退出%s\n", colorGray, interval, colorReset)
	}
}

// renderError 渲染错误信息
func renderError(msg string) {
	fmt.Printf("\n  %s%s错误: %s%s\n\n", colorBold, colorRed, msg, colorReset)
}
