# HydraMancer Runbook

## Overview

HydraMancer is the creator onboarding portal for Experiencenet, serving a landing page at hydramancer.com. It runs on the dashboard server (78.47.174.83) on port 8087 behind hydrareverseproxy.

## Start / Stop / Restart

```bash
sudo systemctl start hydramancer
sudo systemctl stop hydramancer
sudo systemctl restart hydramancer
sudo systemctl status hydramancer
```

## Logs

```bash
journalctl -u hydramancer -f          # follow logs
journalctl -u hydramancer --since today
```

## Health Check

```bash
curl http://localhost:8087/api/v1/health
curl https://hydramancer.com/api/v1/health
```

Expected response: `{"status":"ok"}`

## Configuration

Config file: `~/.hydramancer/config.yaml`

```yaml
server:
  domain: hydramancer.com
```

If no config file exists, defaults are used (domain: hydramancer.com).

## Deployment

Binary is auto-updated from releases.experiencenet.com every 6 hours via hydrarelease. The systemd service restarts automatically after update.

Manual deployment:
1. Build: `make build`
2. Copy binary to server: `scp bin/hydramancer root@78.47.174.83:/usr/local/bin/`
3. Restart: `ssh root@78.47.174.83 systemctl restart hydramancer`

## Release

Tag a version to trigger GitHub Actions:
```bash
git tag v0.1.0
git push origin v0.1.0
```

The workflow builds multi-platform binaries, publishes to releases.experiencenet.com, and creates a GitHub release.

## Infrastructure

- **Server**: 78.47.174.83 (dashboard server)
- **Port**: 8087
- **Reverse proxy**: hydrareverseproxy routes hydramancer.com -> localhost:8087
- **DNS**: A record hydramancer.com -> 78.47.174.83
- **Systemd**: hydramancer.service

## Troubleshooting

### Service won't start
1. Check logs: `journalctl -u hydramancer -n 50`
2. Verify binary exists: `ls -la /root/hydramancer` or `/usr/local/bin/hydramancer`
3. Check port conflict: `ss -tlnp | grep 8087`

### Site unreachable
1. Check service: `systemctl status hydramancer`
2. Check reverse proxy: `systemctl status hydrareverseproxy`
3. Check DNS: `dig hydramancer.com`
4. Check local: `curl http://localhost:8087/api/v1/health`

### Auto-update not working
1. Check logs for "Auto-update" messages
2. Verify releases.experiencenet.com is accessible from the server
3. Ensure service is not running in dev mode (--dev disables auto-update)
