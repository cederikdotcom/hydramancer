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

## Delivery paths

The portal exposes two creator front doors on the landing page:

- **Ship a service** (`/deploy`) — containerized HTTP services onto Hydra.
- **Publish an experience** (`/experience`) — Unreal builds through the
  experience lifecycle (draft &rarr; staging &rarr; live).

The `/experience` page is the creator-facing home for the build-delivery modes
described below. It opens with three entry points — in the browser, from the
command line, or automated through Perforce — then presents one lifecycle with
the two by-hand modes plus the Perforce mode, mirroring the `/deploy` quickstart
layout. The delivery-mode cards that used to live on the landing page now live on
`/experience`. Keep the page consistent with
`hydraexperiencelibrary/docs/getting-started.md` (the source of truth for the
commands).

The web-interface entry point: packaging via `hydraunrealengine serve` (local web
UI on :9100) and the ExperienceLibrary admin at
`hydraexperiencelibrary.experiencenet.com/admin`, which drives the lifecycle
(promote / rollback / pause / resume / retire — the confirmed web routes). Note
`create` and the first `stage` have no web route; they stay CLI (mode A) or come
from the Perforce auto-stage (mode B).

Creators deliver builds by one of two modes. Pick the mode per agency.

1. **HydraUnrealEngine to the release server** (mode A, manual, default). The
   creator packages a project with HydraUnrealEngine, then rsyncs the packaged
   output to `releases.experiencenet.com:/var/www/releases/builds/<name>/<version>/`
   and registers/stages/promotes it in HydraExperienceLibrary. This rsync-to-releases
   path is the canonical manual delivery method. Use it for one-off or occasional
   deliveries and for a first build.

2. **Perforce** (mode B, automated, for agencies that submit often). The creator
   submits packaged builds to a stream depot. HydraPerforce detects the changelist
   and fires a build-notify webhook (`POST /api/v1/builds/notify`) to
   ExperienceLibrary, which auto-stages the experience whose `--watch` target
   matches. No manual upload; `promote` is the only manual gate.

The Perforce path is live for **Cyborn / Gallo-Romeins Museum**:

| | |
|---|---|
| Dev getting-started | [`docs/onboarding/galloromeins-perforce-getting-started.md`](../onboarding/galloromeins-perforce-getting-started.md) |
| Server + watcher ops | `hydraperforcewatcher/docs/runbooks/galloromeins-perforce.md` |
| Server | `ssl:perforce.galloromeins.experiencenet.com:1666`, depot `//galloromeinsmuseum/main` |
| Tracker | issues.experiencenet.com #479 (setup), #480 (Koen onboarding) |

Send a new Cyborn developer the getting-started doc. To create their account,
see the admin appendix in that same doc.

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
provision:
  perforce_url: "http://195.201.88.170:8090"   # hydraperforceprovision
```

The container carries no config file, so set the provisioning URL by env instead
(it survives an image rebuild, unlike a file baked into the rootfs):

```bash
incus config set hydramancer environment.HYDRAMANCER_PROVISION_PERFORCE_URL http://195.201.88.170:8090
incus restart hydramancer
```

## Provisioning wiring

`POST /api/v1/provision/perforce` proxies a creator's Perforce access request to
**hydraperforceprovision**, forwarding their iamnim session
(`iamnim_session` cookie, `X-Iamnim-Session` header, or `?token=`). The portal
holds no credentials and does no validation: hydraperforceprovision checks the
session against iamnim, confirms org membership, and mints the depot/account.

Responses: `503` (URL not configured), `401` (no session), `502` (provisioning
service unreachable), otherwise the upstream status/body pass straight through.

Network path: hydraperforceprovision listens on `195.201.88.170:8090`, UFW-limited
to this portal node's egress IP (`94.224.39.25`), and is iamnim-session-gated.
The venue egress IP can change; the durable fix is to reach it over the WireGuard
mesh instead. A login UX on `/experience` (redirect to iamnim, then call this
route) is still to build — the proxy is ready for it.

## Deployment

This runs as an **OCI container**, so the image is the unit of deployment:

```
oci.entrypoint: hydramancer serve --dev --addr :8080
image.id:       hydramancer:v<X.Y.Z>   (scaleregistry.experiencenet.com)
```

Tagging a release both publishes binaries to `releases.experiencenet.com` and
pushes an OCI image to the scale registry. Deploying means moving the container
to the new image. Copying a binary into the container is a hotfix and is **lost
on any container recreate**, because the image still carries the old one.

### Why auto-update is off here

`serve` wires `updater.StartAutoCheck(6*time.Hour, true)`, but it is skipped in
this deployment for two independent reasons, both deliberate:

- the container sets `environment.HYDRA_AUTO_UPDATE: 'off'`
- the entrypoint passes `--dev`, and auto-update is gated on `!dev`

That is correct for a container: the updater replaces a binary on disk and
restarts a systemd unit, and there is no systemd here (PID 1 is `incusd`,
hydramancer runs as a child). Do not "fix" this by removing the guards.

### Emergency binary update

`update` and `check-update` exist for host installs and for forcing a binary
swap when redeploying the image is not practical. `--force` skips the prompt so
it works over `hydracluster exec`:

```bash
$HYDRACLUSTER_BIN exec node-50ab5309 \
  "/usr/bin/incus exec hydramancer -- /usr/local/bin/hydramancer update --force" \
  --server "$S" --admin-token "$TOKEN"
$HYDRACLUSTER_BIN exec node-50ab5309 "/usr/bin/incus restart hydramancer" \
  --server "$S" --admin-token "$TOKEN"
```

The restart step is required: the updater's own `systemctl restart` cannot work
in this container, so it warns and leaves the old process running the old
binary. The previous binary is kept at `/usr/local/bin/hydramancer.backup`
(mode 0644 by design; `chmod +x` it to roll back).

Note the Pi fleet is **arm64**.

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

Expected. Auto-update is deliberately off in the container, so a release does
not deploy itself. Follow Deployment above.

### Version reverted after a container recreate

Expected if the last deploy was an emergency binary swap rather than an image
update. The image still carries the old binary. Redeploy the image.
