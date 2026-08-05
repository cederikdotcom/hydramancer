# HydraMancer Runbook

## Overview

HydraMancer is the creator onboarding portal for Experiencenet, serving a landing page at **hydramancer.experiencenet.com**. It runs as an Incus container on the Pi fleet, fronted by hydrascalerouter.

> `hydramancer.com` is not ours and does not resolve. Earlier revisions of this runbook used it throughout; the domain is `hydramancer.experiencenet.com`.

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

## Infrastructure

| | |
|---|---|
| Domain | `hydramancer.experiencenet.com` |
| Public IP | 141.227.136.199 (hydraguard-brussels, node-2d5fba78) |
| Ingress | hydrascalerouter + Traefik on hydraguard-brussels |
| Backend | `http://10.0.0.5:30007` |
| Host | pi-node-004-nvme (node-50ab5309) |
| Runtime | Incus container `hydramancer`, PID 1 is `incusd` |
| Binary | `/usr/local/bin/hydramancer` inside the container |

The container has **no systemd and no journald**. Any instruction using
`systemctl` or `journalctl` against this service is wrong.

## Operations

All access is via `hydracluster exec` against the host node, then `incus exec`
into the container. See [[feedback_hydra_api_only]]: no direct SSH.

```bash
HYDRACLUSTER_BIN=/home/claude-user/hydracluster/bin/hydracluster
S=https://hydracluster.experiencenet.com
$HYDRACLUSTER_BIN exec node-50ab5309 "/usr/bin/incus list -c ns4 -f csv" --server "$S" --admin-token "$TOKEN"
```

### Restart

```bash
$HYDRACLUSTER_BIN exec node-50ab5309 "/usr/bin/incus restart hydramancer" --server "$S" --admin-token "$TOKEN"
```

### Version

```bash
$HYDRACLUSTER_BIN exec node-50ab5309 "/usr/bin/incus exec hydramancer -- /usr/local/bin/hydramancer version" --server "$S" --admin-token "$TOKEN"
```

### Health check

```bash
curl https://hydramancer.experiencenet.com/api/v1/health     # {"status":"ok"}
```

## Configuration

Config file inside the container: `~/.hydramancer/config.yaml`

```yaml
server:
  domain: hydramancer.experiencenet.com
```

## Deployment

**This binary has no `update` command.** It does not self-update, and it is not
on a 6-hour auto-update cycle; earlier revisions of this runbook claimed both.
Tagging a release publishes the binary but does not deploy it.

Deploy by fetching the released binary into the container and restarting it. The
release server serves versioned paths only, so use the explicit version. There is
no `production/latest/` alias for this project:

```bash
# inside the container
curl -fsSL -o /tmp/hydramancer.new \
  https://releases.experiencenet.com/hydramancer/production/v<X.Y.Z>/hydramancer-linux-arm64
chmod +x /tmp/hydramancer.new
cp /usr/local/bin/hydramancer /usr/local/bin/hydramancer.backup
mv /tmp/hydramancer.new /usr/local/bin/hydramancer
# then, on the host
incus restart hydramancer
```

Note the architecture: the Pi fleet is **arm64**.

Adding the standard `hydrarelease/pkg/updater` update command to this binary
would remove the manual step and make it match the other hydra services.

## Release

Tag a version to trigger GitHub Actions:

```bash
git -C <repo> tag v<X.Y.Z>
git -C <repo> push origin v<X.Y.Z>
```

The workflow builds multi-platform binaries, publishes them to
`releases.experiencenet.com`, pushes an OCI image to the scale registry, and
creates a GitHub release. Deployment is still manual, see above.

## Troubleshooting

### Site unreachable

1. Health: `curl https://hydramancer.experiencenet.com/api/v1/health`
2. Container running? `incus list` on node-50ab5309, expect `hydramancer RUNNING`
3. Backend listening? `ss -tlnp | grep 30007` on the host, expect `incusd`
4. Route registered? `grep -A4 hydramancer /var/lib/hydrascalerouter/routes.json`
   on node-2d5fba78
5. DNS: `getent hosts hydramancer.experiencenet.com`, expect 141.227.136.199

### Wrong version serving after a release

Expected: the binary does not self-update. Follow Deployment above.
