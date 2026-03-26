package main

import (
	"regexp"
	"sort"
	"strings"
)

// =============================================================================
// 区域分组与分类
// =============================================================================

var cleanNameRe = regexp.MustCompile(`-BGP-|-NF-|-NF$`)

// regionOrder 区域显示顺序
var regionOrder = []string{"中国香港", "新加坡", "中国台湾", "日本", "美国", "其他地区"}

// regionMeta 区域元信息
var regionMeta = map[string]struct {
	Flag   string
	NameEN string
}{
	"中国香港": {"🇭🇰", "Hong Kong"},
	"新加坡":   {"🇸🇬", "Singapore"},
	"中国台湾": {"🇨🇳", "Taiwan"},
	"日本":     {"🇯🇵", "Japan"},
	"美国":     {"🇺🇸", "United States"},
	"其他地区": {"🌐", "Other"},
}

// getRegion 根据节点名称判断区域
func getRegion(name string) string {
	if strings.Contains(name, "SG") || strings.Contains(name, "🇸🇬") {
		return "新加坡"
	}
	if strings.Contains(name, "HK") || strings.Contains(name, "🇭🇰") {
		return "中国香港"
	}
	if strings.Contains(name, "TW") || strings.Contains(name, "🇹🇼") || strings.Contains(name, "🇨🇳") {
		// 订阅中台湾节点用 🇨🇳 且名称含 TW
		if strings.Contains(name, "TW") {
			return "中国台湾"
		}
		// 仅有 🇨🇳 而无 TW，也归入台湾（根据订阅数据特征）
		return "中国台湾"
	}
	if strings.Contains(name, "JP") || strings.Contains(name, "🇯🇵") {
		return "日本"
	}
	if strings.Contains(name, "US") || strings.Contains(name, "🇺🇸") || strings.Contains(name, "🇺🇲") {
		return "美国"
	}
	return "其他地区"
}

// cleanNodeName 清洗节点名称
func cleanNodeName(name string) string {
	// 去除 emoji 前缀
	name = strings.TrimSpace(name)
	// 去除 -BGP-、-NF- 等标记
	name = cleanNameRe.ReplaceAllString(name, "-")
	// 去除首尾的 -
	name = strings.Trim(name, "-")
	return name
}

// groupByRegion 将检测结果按区域分组
func groupByRegion(results []CheckResult) []RegionData {
	regionMap := make(map[string][]CheckResult)

	for _, r := range results {
		region := getRegion(r.Node.Name)
		regionMap[region] = append(regionMap[region], r)
	}

	var regions []RegionData
	for _, regionName := range regionOrder {
		items, ok := regionMap[regionName]
		if !ok || len(items) == 0 {
			continue
		}

		// 排序：正常节点延迟升序，故障节点置底
		sort.Slice(items, func(i, j int) bool {
			if items[i].Category == "fault" && items[j].Category != "fault" {
				return false
			}
			if items[i].Category != "fault" && items[j].Category == "fault" {
				return true
			}
			return items[i].TCPPing < items[j].TCPPing
		})

		// 统计
		stats := NodeStats{Total: len(items)}
		for _, item := range items {
			switch item.Category {
			case "fast":
				stats.Fast++
			case "normal":
				stats.Normal++
			case "high_latency":
				stats.HighLatency++
			case "fault":
				stats.Fault++
			}
		}

		meta := regionMeta[regionName]
		regions = append(regions, RegionData{
			Name:    regionName,
			Flag:    meta.Flag,
			NameEN:  meta.NameEN,
			Results: items,
			Stats:   stats,
		})
	}

	return regions
}

// calcTotalStats 计算总体统计
func calcTotalStats(regions []RegionData) NodeStats {
	var total NodeStats
	for _, r := range regions {
		total.Total += r.Stats.Total
		total.Fast += r.Stats.Fast
		total.Normal += r.Stats.Normal
		total.HighLatency += r.Stats.HighLatency
		total.Fault += r.Stats.Fault
	}
	return total
}
