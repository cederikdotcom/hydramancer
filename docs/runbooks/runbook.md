# HydraMancer Runbook

## Overview

HydraMancer is the creator onboarding portal for Experiencenet, serving a landing page at hydramancer.com. It runs on the dashboard server (78.47.174.83) on port 8087 behind hydrareverseproxy.

## Creator toolchain

HydraMancer is the front door for external creators (agencies and contractors
building experiences for Experiencenet). The tool it points them at is
**HydraUnrealEngine**, a Windows CLI that packages Unreal projects, runs
preflight checks, and uploads builds to HydraTransfer.

| | |
|---|---|
| Repo | [`hydraunrealengine`](https://github.com/cederikdotcom/hydraunrealengine) |
| Runbook | `hydraunrealengine/docs/runbooks/runbook.md` |
| Job dashboard | `https://hydraunrealengine.experiencenet.com` (repo: `hydraunrealengine-server`) |
| Platform | Windows only, requires Unreal Engine installed locally |

### What to send a creator

There is no download page and no hosted installer. `…/hydraunrealengine/` and
`…/hydraunrealengine/install.ps1` both 404. Send this PowerShell line:

```powershell
Invoke-WebRequest -Uri 'https://releases.experiencenet.com/hydraunrealengine/production/latest/hydraunrealengine-windows-amd64.exe' -OutFile hydraunrealengine.exe
```

Then `hydraunrealengine detect` reports what engine and toolchain their machine
sees, and `hydraunrealengine serve` gives them a local web UI on port 9100 if
they would rather not use the CLI.

Keep the landing page's download command in sync with the runbook above. If a
download page or published installer ever lands, update both.

### Scope note

HydraUnrealEngine packages a project that already builds. It is not a
troubleshooting tool for a project that will not open, and the server side is a
job tracker with no build capacity. Getting a broken project to open still needs
a Windows machine with the right Unreal version on it.

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
