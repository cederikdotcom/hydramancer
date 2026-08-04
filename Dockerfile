# hydramancer as an OCI application container — a single static binary, one process.
#
# Runs in --dev mode: plain HTTP on :8080, with TLS terminated upstream by
# hydrascalerouter. The default serve binds 80/443 and manages certificates
# itself, which is wrong inside a scale.
FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG VERSION=dev
# CGO off so the binary is fully static and needs nothing from the final image.
RUN CGO_ENABLED=0 go build -trimpath \
      -ldflags "-s -w -X github.com/cederikdotcom/hydramancer/internal/cli.Version=${VERSION}" \
      -o /hydramancer ./cmd/hydramancer

FROM alpine:3.21
# ca-certificates only: the binary is static, but its outbound calls are HTTPS.
RUN apk add --no-cache ca-certificates
COPY --from=build /hydramancer /usr/local/bin/hydramancer

# Honoured by a guard in serve.go — the ENV alone does nothing. The updater
# rewrites the binary and restarts a systemd unit that does not exist here.
ENV HYDRA_AUTO_UPDATE=off

# config.yaml lives at /root/.hydramancer/config.yaml. The instance retired on
# 2026-08-04 had NO config and NO state directory at all — it ran on defaults,
# hand-deployed from /root and never routed. Attach an Incus disk device on
# /root/.hydramancer if it ever gains state.
#
# Deliberately NO VOLUME declaration: Incus's OCI runtime cannot satisfy an
# anonymous volume and the container fails to start with
#   Failed to mount "none" onto ".../rootfs/run"
EXPOSE 8080

ENTRYPOINT ["hydramancer"]
CMD ["serve", "--dev", "--addr", ":8080"]
