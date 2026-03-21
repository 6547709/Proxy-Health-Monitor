#!/usr/bin/env python3
import json
import sys
from typing import Any, Dict, List, Tuple


def load_data() -> Dict[str, Any]:
    if len(sys.argv) > 1:
        with open(sys.argv[1], 'r', encoding='utf-8') as f:
            return json.load(f)
    return json.load(sys.stdin)


def health_status(stats: Dict[str, int]) -> Tuple[str, str, float]:
    total = stats.get('total', 0) or 0
    fast = stats.get('fast', 0) or 0
    normal = stats.get('normal', 0) or 0
    available = fast + normal
    if total <= 0:
        return '⚪ 未知', '没有检测到有效节点数据', 0.0
    availability = available / total
    if availability >= 0.6:
        return '🟢 健康', '当前节点池整体可用性较好', availability
    if availability >= 0.3:
        return '🟡 一般', '当前节点池可用，但质量波动较明显', availability
    if availability > 0:
        return '🟠 偏差', '当前有少量可用节点，建议优先挑选低延迟节点使用', availability
    return '🔴 异常', '当前没有 fast / normal 节点，整体健康度较差', availability


def collect_nodes(data: Dict[str, Any]) -> List[Dict[str, Any]]:
    rows: List[Dict[str, Any]] = []
    for region in data.get('regions', []) or []:
        region_name = region.get('name') or region.get('name_en') or '未知区域'
        for node in region.get('nodes', []) or []:
            item = dict(node)
            item['_region_name'] = region_name
            rows.append(item)
    return rows


def node_rank(node: Dict[str, Any]) -> Tuple[int, float, str]:
    category = node.get('category', 'fault')
    latency = node.get('latency')
    if latency is None:
        latency = 999999
    priority = {
        'fast': 0,
        'normal': 1,
        'high_latency': 2,
        'fault': 9,
    }.get(category, 8)
    return (priority, float(latency), str(node.get('name', '')))


def top_healthy_nodes(nodes: List[Dict[str, Any]], limit: int = 3) -> List[Dict[str, Any]]:
    candidates = [
        n for n in nodes
        if n.get('category') in ('fast', 'normal', 'high_latency')
        and n.get('tcp_connected') is True
    ]
    candidates.sort(key=node_rank)
    return candidates[:limit]


def top_fault_regions(data: Dict[str, Any], limit: int = 2) -> List[str]:
    items = []
    for region in data.get('regions', []) or []:
        stats = region.get('stats', {}) or {}
        items.append((stats.get('fault', 0), region.get('name') or region.get('name_en') or '未知区域'))
    items.sort(reverse=True)
    return [name for _, name in items[:limit] if _ > 0]


def build_observations(stats: Dict[str, int], data: Dict[str, Any], top_nodes: List[Dict[str, Any]]) -> List[str]:
    obs: List[str] = []
    fast = stats.get('fast', 0) or 0
    normal = stats.get('normal', 0) or 0
    high_latency = stats.get('high_latency', 0) or 0
    fault = stats.get('fault', 0) or 0
    total = stats.get('total', 0) or 0

    if fast == 0 and normal == 0:
        if high_latency > 0:
            obs.append('当前没有 fast / normal 节点，仅存在高延迟可用节点')
        else:
            obs.append('当前没有可直接推荐的健康节点')

    if total > 0 and fault / total >= 0.7:
        obs.append('故障节点占比过高，更像订阅源或服务端池整体异常，而不是单点抖动')

    fault_regions = top_fault_regions(data)
    if fault_regions:
        obs.append(f"问题较集中的区域：{'、'.join(fault_regions)}")

    if top_nodes:
        best = top_nodes[0]
        obs.append(
            f"当前最优可用节点为 {best.get('name', '未知节点')}（{best.get('_region_name', '未知区域')}，{best.get('latency', '-')} ms）"
        )

    return obs[:4]


def build_actions(status_text: str, top_nodes: List[Dict[str, Any]]) -> List[str]:
    actions: List[str] = []
    if top_nodes:
        actions.append('优先人工验证 Top 3 节点的实际可用性，并作为临时切换候选')
    if status_text.startswith('🔴') or status_text.startswith('🟠'):
        actions.append('抽样检查故障节点的 DNS / TCP 失败模式，确认是否为共性问题')
        actions.append('检查订阅源质量、上游服务端状态，必要时更换或回滚订阅')
    else:
        actions.append('保留当前订阅，同时持续观察区域间的延迟波动')
    return actions[:3]


def print_report(data: Dict[str, Any]) -> None:
    stats = data.get('stats', {}) or {}
    status_text, summary_text, availability = health_status(stats)
    nodes = collect_nodes(data)
    top_nodes = top_healthy_nodes(nodes, 3)
    observations = build_observations(stats, data, top_nodes)
    actions = build_actions(status_text, top_nodes)

    total = stats.get('total', 0) or 0
    fast = stats.get('fast', 0) or 0
    normal = stats.get('normal', 0) or 0
    high_latency = stats.get('high_latency', 0) or 0
    fault = stats.get('fault', 0) or 0
    healthy = fast + normal

    print('代理健康度检查结果')
    print(f'- 整体状态：{status_text}')
    print(f'- 可用率：{availability * 100:.1f}%')
    print(f'- 结论：{summary_text}')
    print()

    print('核心指标')
    print(f'- 总节点：{total}')
    print(f'- 健康节点：{healthy}')
    print(f'- 可用但高延迟：{high_latency}')
    print(f'- 故障节点：{fault}')
    print(f"- 区域数：{len(data.get('regions', []) or [])}")
    print()

    print('Top 3 推荐节点')
    if top_nodes:
        for idx, node in enumerate(top_nodes, 1):
            print(
                f"{idx}. {node.get('name', '未知节点')}｜{node.get('_region_name', '未知区域')}｜{node.get('latency', '-')} ms｜{node.get('category', 'unknown')}"
            )
    else:
        print('- 当前没有可推荐的节点')
    print()

    print('重点观察')
    if observations:
        for item in observations:
            print(f'- {item}')
    else:
        print('- 暂无明显异常特征')
    print()

    print('建议动作')
    for item in actions:
        print(f'- {item}')


if __name__ == '__main__':
    print_report(load_data())
