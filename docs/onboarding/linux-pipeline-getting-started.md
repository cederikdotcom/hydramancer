# Getting started: the Linux pipeline (git push to deploy)

For a Linux/container service, you do not build or launch anything by hand. On a
`v*` tag push to your watched repo, the pipeline builds your `Dockerfile` on a
native arm64 builder and deploys the result as a Hydra scale with automatic
HTTPS on a `*.experiencenet.com` domain.

This is the git equivalent of the manual five steps on the Deploy page - it just
happens on every tag.

## The shape

```
git push v1.2.3
      |
      v
hydragitwatcher   (detects the tag, emits a build-notify)
      |
      v
hydralinuxpipeline  (control plane: tracks the job)
      |
      v
hydrapipelinerunnerlinux  (launches an ephemeral rootless-buildkit scale on a
      |                     hydraskin node, builds your Dockerfile natively,
      |                     pushes the image, tears the builder down)
      v
deploy  (hydraskin launch on the target node over the exec transport)
      |
      v
https://<your-scale>.experiencenet.com
```

- Watchers are keyed by version control (git here), pipelines and builders by
  platform (Linux/container here). A different VCS gets its own watcher; a
  different platform gets its own pipeline and builder.
- The build is **native** (arm64 on arm64 hardware), never emulated.
- Scales are their own vertical: delivery is a scale-registry push plus a
  `hydraskin launch`, never routed through HydraRelease or HydraExperienceLibrary.

## What your repo needs

1. A `Dockerfile` that builds a container serving **plain HTTP on :8080** (Hydra
   terminates TLS at the edge). No `VOLUME`; put state under `/data` and it is
   attached as a disk device at launch.
2. A tag push. A `v*` tag deploys; a branch push builds the image only (CI, no
   deploy).
3. Optional `.hydrabuild.yaml` at the repo root to refine the scale intent
   (name, domain, port, health path, state disk).

## What you get told

The pipeline exposes the job through `GET /api/v1/builds` and a live SSE stream;
each build moves through `queued -> building -> pushed -> deploying -> live` (or
`failed` with the reason). The runner exposes the builder lifecycle through
`GET /api/v1/builders`.

## Reference

- Pipeline: `cederikdotcom/hydralinuxpipeline` - README plus
  `docs/runbooks/runbook.md` (also served at `/api/v1/runbook`).
- Builder manager: `cederikdotcom/hydrapipelinerunnerlinux` -
  `docs/runbooks/runbook.md`.
- Watcher: `cederikdotcom/hydragitwatcher` (set `pipeline.url` for notify-only
  mode, which feeds the pipeline instead of building itself).
- Tracking: issue #508 on issues.experiencenet.com.

## Status

Today the pipeline is **operator-onboarded**: to have your repo watched and a
domain assigned, ask the platform team. Self-service enrolment through this
portal is the next step.
