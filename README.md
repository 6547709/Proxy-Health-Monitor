# 🌐 代理节点健康监控 (Proxy Health Monitor)

终端 CLI 工具，从订阅链接获取代理节点，在本地执行健康检测（DNS 解析 → TCP 连通 → 延迟测量），以彩色终端界面展示结果。

## 快速开始

```bash
# 方式一：直接使用命令行参数
./proxy-monitor -url "https://sub.example.com/link/xxx?clash=2"

# 方式二：使用环境变量
export SUBSCRIBE_URL="https://sub.example.com/link/xxx?clash=2"
./proxy-monitor
```

## 下载

前往 [Releases](../../releases) 下载对应平台的二进制文件：

| 平台 | 文件名 |
|------|--------|
| Linux x64 | `proxy-monitor-linux-amd64` |
| Linux ARM64 | `proxy-monitor-linux-arm64` |
| macOS x64 | `proxy-monitor-darwin-amd64` |
| macOS Apple Silicon | `proxy-monitor-darwin-arm64` |
| Windows x64 | `proxy-monitor-windows-amd64.exe` |

下载后赋予执行权限：
```bash
chmod +x proxy-monitor-*
```

## 配置选项

所有参数同时支持**命令行参数**和**环境变量**（CLI 参数优先级更高）。

| CLI 参数 | 环境变量 | 默认值 | 说明 |
|----------|----------|--------|------|
| `-url` | `SUBSCRIBE_URL` | (必填) | 订阅地址 |
| `-interval` | `INTERVAL` | `60` | 刷新间隔（秒） |
| `-concurrent` | `CONCURRENT` | `20` | 最大并发检测数 |
| `-timeout` | `TIMEOUT` | `5` | 单节点超时（秒） |
| `-once` | - | `false` | 仅执行一次检测 |
| `-no-color` | `NO_COLOR` | `false` | 禁用颜色输出 |
| `-version` | - | - | 显示版本号 |

### 示例

```bash
# 持续监控，每 30 秒刷新
./proxy-monitor -url "https://..." -interval 30

# 仅检测一次
./proxy-monitor -url "https://..." -once

# 使用环境变量运行
export SUBSCRIBE_URL="https://..."
export INTERVAL=30
./proxy-monitor

# Docker / CI 场景
SUBSCRIBE_URL="https://..." INTERVAL=120 ./proxy-monitor -once
```

## 检测逻辑

对每个代理节点执行三步检测：

1. **DNS 解析** — 检查域名是否能解析为 IP 地址
2. **TCP 连通** — 使用解析到的 IP + 端口尝试 TCP 连接
3. **延迟测量** — TCP 连接建立耗时即为延迟

### 延迟分类

| 分类 | 延迟范围 | 图标 |
|------|----------|------|
| 高速 | < 150ms | 🚀 |
| 正常 | 150-220ms | ✅ |
| 高延迟 | 220-310ms | ⚠️ |
| 故障 | > 310ms 或超时 | ❌ |

### 区域分组

节点根据名称自动分组：🇭🇰 中国香港、🇸🇬 新加坡、🇨🇳 中国台湾、🇯🇵 日本、🇺🇸 美国、🌐 其他地区

## 从源码编译

```bash
git clone https://github.com/your-username/Proxy-Health-Monitor.git
cd Proxy-Health-Monitor
go build -o proxy-monitor .
```

## 自动发布

Push tag 到 GitHub 即可触发自动编译和发布：

```bash
git tag v2.0.0
git push origin v2.0.0
```

GitHub Actions 会自动交叉编译 5 个平台的二进制并创建 Release。

## License

MIT
