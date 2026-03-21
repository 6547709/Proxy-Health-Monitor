# Proxy Health Check Usage

## Required runtime inputs

### Preferred
- `SUBSCRIBE_URL`: Clash-format subscription URL

### Optional
- `PROXY_MONITOR_BIN`: full path to `proxy-monitor` binary

If `PROXY_MONITOR_BIN` is not set, the wrapper expects the binary at:

- `bin/proxy-monitor` relative to this skill directory

## Recommended operator workflow

1. Ensure `proxy-monitor` exists and is executable
2. Export `SUBSCRIBE_URL`
3. Run `scripts/run_proxy_health.sh`
4. Pipe JSON to `scripts/summarize_proxy_health.py`

## Example

```bash
export SUBSCRIBE_URL='https://sub.example.com/link/xxx?clash=2'
export PROXY_MONITOR_BIN='/absolute/path/to/proxy-monitor'

./scripts/run_proxy_health.sh | python3 ./scripts/summarize_proxy_health.py
```

## Notes

- Default interaction should be one-shot, not watch mode
- Avoid printing full subscription URLs in shared contexts
- Use watch mode only for explicit continuous monitoring needs
- The measured result reflects health from the current machine and network path

## Future extensions

- Add a script to emit Markdown tables or Feishu-ready summaries
- Add history persistence for trend comparison
- Add threshold-based alerts for failed node ratio or regional degradation
- Add support for multiple subscription profiles
