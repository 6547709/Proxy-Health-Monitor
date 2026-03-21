---
description: 使用 proxy-monitor 工具检测代理服务器节点的健康状态
---

# proxy-monitor - 代理节点健康监控工具

## 工具概述

`proxy-monitor` 是一个 CLI 工具，从订阅链接获取代理服务器节点列表，在本地执行健康检测（DNS 解析 → TCP 连通 → 代理延迟），输出检测结果。

## 快速调用

```bash
# 执行一次检测，输出终端带颜色的结果
proxy-monitor -url "https://sub.example.com/link/xxx?clash=2"

# 执行一次检测，输出 JSON 格式（推荐程序调用使用此方式）
proxy-monitor -url "https://sub.example.com/link/xxx?clash=2" -json

# 使用环境变量
SUBSCRIBE_URL="https://..." proxy-monitor -json
```

## 参数说明

| 参数 | 环境变量 | 类型 | 默认值 | 说明 |
|------|----------|------|--------|------|
| `-url` | `SUBSCRIBE_URL` | string | (必填) | Clash 格式订阅地址 |
| `-json` | `JSON_OUTPUT` | bool | false | JSON 格式输出，适合程序解析 |
| `-watch` | - | bool | false | 持续监控模式（默认执行一次） |
| `-interval` | `INTERVAL` | int | 60 | 持续模式刷新间隔（秒） |
| `-concurrent` | `CONCURRENT` | int | 20 | 最大并发检测数 |
| `-timeout` | `TIMEOUT` | int | 10 | 单节点超时（秒） |
| `-no-color` | `NO_COLOR` | bool | false | 禁用颜色输出 |
| `-version` | - | - | - | 显示版本号 |

**优先级**: CLI 参数 > 环境变量 > 默认值

## JSON 输出格式

使用 `-json` 参数时，输出以下 JSON 结构：

```json
{
  "version": "2.1.0",
  "update_time": "2026-03-21T09:30:00+08:00",
  "stats": {
    "Total": 96,
    "Fast": 20,
    "Normal": 15,
    "HighLatency": 30,
    "Fault": 31
  },
  "regions": [
    {
      "name": "中国香港",
      "flag": "🇭🇰",
      "name_en": "Hong Kong",
      "stats": { "Total": 35, "Fast": 8, "Normal": 5, "HighLatency": 12, "Fault": 10 },
      "nodes": [
        {
          "name": "V301U-1X-HK",
          "server": "xxx.cache872671.com",
          "port": 1301,
          "dns_resolved": true,
          "resolved_ips": ["103.181.165.71"],
          "tcp_connected": true,
          "latency": 210,
          "category": "normal"
        },
        {
          "name": "V302U-1X-HK",
          "server": "xxx.cache872671.com",
          "port": 1302,
          "dns_resolved": true,
          "tcp_connected": false,
          "latency": 0,
          "category": "fault",
          "error": "代理测试失败: ..."
        }
      ]
    }
  ]
}
```

### 字段说明

- **category**: `fast`(<150ms) / `normal`(150-220ms) / `high_latency`(220-310ms) / `fault`(>310ms或失败)
- **latency**: 通过代理访问 `http://www.gstatic.com/generate_204` 的往返延迟(ms)，与 Clash 测量方式一致
- **error**: 仅在检测失败时出现

## 检测流程

1. **DNS 解析**: 检查代理服务器域名是否可解析
2. **TCP 连通**: 检查是否可以连接到代理服务器端口
3. **代理延迟**: 通过代理协议建立隧道，访问 Google generate_204 测量端到端延迟

## 使用示例

### 场景1: 获取所有节点状态（JSON）
```bash
proxy-monitor -url "https://sub.ssr.sh/link/xxx?clash=2" -json
```

### 场景2: 只看某个区域的节点
```bash
proxy-monitor -url "https://..." -json | jq '.regions[] | select(.name == "中国香港")'
```

### 场景3: 找出所有高速节点
```bash
proxy-monitor -url "https://..." -json | jq '[.regions[].nodes[] | select(.category == "fast")]'
```

### 场景4: 快速检测（减少超时）
```bash
proxy-monitor -url "https://..." -json -timeout 5 -concurrent 30
```
