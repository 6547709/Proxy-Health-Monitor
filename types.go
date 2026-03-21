package main

// =============================================================================
// 类型定义
// =============================================================================

// ProxyNode 从订阅YAML中解析出的代理节点
type ProxyNode struct {
	Name      string         `yaml:"name"`
	Type      string         `yaml:"type"`
	Server    string         `yaml:"server"`
	Port      int            `yaml:"port"`
	RawConfig map[string]any // 原始完整配置，用于创建代理适配器
}

// CheckResult 单个节点的检测结果
type CheckResult struct {
	Node         ProxyNode
	DNSResolved  bool     // DNS 是否解析成功
	ResolvedIPs  []string // 解析到的 IP 列表
	TCPConnected bool     // TCP 是否连通
	Latency      int      // 延迟 ms（TCP连接耗时）
	Category     string   // fast / normal / high_latency / fault
	Error        string   // 错误信息（如有）
}

// NodeStats 节点统计
type NodeStats struct {
	Total       int `json:"total"`
	Fast        int `json:"fast"`
	Normal      int `json:"normal"`
	HighLatency int `json:"high_latency"`
	Fault       int `json:"fault"`
}

// RegionData 区域数据
type RegionData struct {
	Name    string
	Flag    string
	NameEN  string
	Results []CheckResult
	Stats   NodeStats
}

// AppConfig 应用配置（支持 CLI 参数 + 环境变量）
type AppConfig struct {
	SubscriptionURL string // 订阅地址
	Interval        int    // 刷新间隔（秒）
	Concurrent      int    // 最大并发检测数
	Timeout         int    // 单节点超时（秒）
	Watch           bool   // 持续监控模式（默认 false，即执行一次）
	NoColor         bool   // 禁用颜色输出
	JSONOutput      bool   // JSON 格式输出（适合程序调用）
}

// ClashConfig 订阅YAML的顶层结构
type ClashConfig struct {
	Proxies []ProxyNode `yaml:"proxies"`
}
