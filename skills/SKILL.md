---
name: proxy-health-check
description: Check proxy subscription health, node availability, latency, failures, and regional quality using the local proxy-monitor binary. Use when the user asks to inspect代理健康度、代理节点可用性、测速、延迟、故障、最快节点、区域质量，or wants a summary of proxy subscription status from a SUBSCRIBE_URL environment variable.
---

# Proxy Health Check

Use this skill to assess the health of a proxy subscription with the local `proxy-monitor` binary.

## What this skill does

- Run one-shot health checks against a Clash-format subscription URL
- Read the subscription URL from the `SUBSCRIBE_URL` environment variable by default
- Prefer JSON output for deterministic parsing
- Summarize results into a concise operator-friendly report
- Highlight failed nodes, high-latency nodes, fastest nodes, and regional distribution

## Preconditions

Before running a health check, verify these conditions:

1. The local `proxy-monitor` binary is installed and executable
2. `SUBSCRIBE_URL` is set in the environment, unless the user explicitly provides a temporary URL for this run
3. Prefer one-shot execution; do not use `-watch` unless the user explicitly asks for continuous monitoring

## Default execution flow

1. Check whether `SUBSCRIBE_URL` is present
2. Run `scripts/run_proxy_health.sh`
3. Parse and summarize JSON output with `scripts/summarize_proxy_health.py`
4. Return a concise summary first
5. If the user asks for more detail, include regional stats, failed nodes, and fastest nodes

## Output style

Default to this structure:

- Overall status: healthy / mixed / degraded
- Total nodes
- Healthy vs failed nodes
- Fastest nodes (Top 3 to Top 5)
- Regions with the best quality
- Regions or nodes with failures/high latency
- Recommended next action

## Troubleshooting

If the check fails:

- If `SUBSCRIBE_URL` is missing, ask the user to provide or export it
- If `proxy-monitor` is missing, ask the user to install/download the release binary first
- If the subscription cannot be fetched, report it as a subscription or network failure
- If JSON parsing fails, show the raw stderr/stdout snippet and suggest verifying binary version

## Security and privacy

- Treat subscription URLs as sensitive secrets
- Do not write subscription URLs into SKILL.md or reference files
- Prefer environment variables or a local untracked config file
- Avoid echoing the full subscription URL back to the user unless they explicitly ask

## Bundled resources

- `scripts/run_proxy_health.sh`: wrapper for invoking `proxy-monitor`
- `scripts/summarize_proxy_health.py`: converts JSON output into a readable health summary
- `references/usage.md`: usage notes, expected environment variables, and extension ideas
