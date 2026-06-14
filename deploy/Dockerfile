# syntax=docker/dockerfile:1.7
#
# Single multi-stage build for the Somra all-in-one image.
#
# Stages:
#   web-builder : Node + pnpm — produces web/dist (SPA static bundle)
#   go-builder  : Go (CGO disabled) — produces /build/bin/somra
#   runtime     : Alpine + ffmpeg/ffprobe + tini — minimal runtime user
#
# Build:
#   docker build -f deploy/Dockerfile -t somra:dev .
#
# Multi-arch:
#   docker buildx build --platform linux/amd64,linux/arm64 \
#       -f deploy/Dockerfile -t ghcr.io/somralab/somra-media:dev --push .

# ---------------------------------------------------------------------------
# Stage 1: web-builder
# ---------------------------------------------------------------------------
FROM --platform=$BUILDPLATFORM node:22-alpine AS web-builder
WORKDIR /build/web

RUN corepack enable && corepack prepare pnpm@10.4.0 --activate

COPY web/package.json web/pnpm-lock.yaml ./
RUN --mount=type=cache,id=pnpm-store,target=/pnpm-store \
    pnpm config set store-dir /pnpm-store && \
    pnpm install --frozen-lockfile

COPY web/ ./
RUN pnpm build

# ---------------------------------------------------------------------------
# Stage 2: go-builder
# ---------------------------------------------------------------------------
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS go-builder
WORKDIR /build

RUN apk add --no-cache git ca-certificates

# Cache module downloads independently from sources.
COPY go.mod go.sum ./
RUN --mount=type=cache,id=gomodcache,target=/go/pkg/mod \
    go mod download

# Sources after deps so dep changes invalidate fewer layers.
COPY . .

# Build flags:
#   - CGO_ENABLED=0   : pure Go (modernc.org/sqlite is pure Go).
#   - -trimpath       : reproducible paths.
#   - -s -w           : strip debug info to keep the image small.
#   - ldflags inject  : version + commit + builtAt for /api/v1/version.
ARG TARGETOS=linux
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILT_AT=unknown
ENV CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH
RUN --mount=type=cache,id=gomodcache,target=/go/pkg/mod \
    --mount=type=cache,id=gobuildcache,target=/root/.cache/go-build \
    go build -trimpath \
      -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.builtAt=${BUILT_AT}" \
      -o /build/bin/somra ./cmd/somra

# ---------------------------------------------------------------------------
# Stage 3: runtime
# ---------------------------------------------------------------------------
FROM alpine:3.21 AS runtime

# ffmpeg HW stack: libva + Intel media driver + Mesa VAAPI for AMD.
# NVIDIA encode requires host driver passthrough (see deploy/docker-compose.nvidia.yml).
RUN apk add --no-cache \
      ffmpeg \
      ffmpeg-libs \
      intel-media-driver \
      libva \
      mesa-va-drivers \
      tini \
      ca-certificates \
      tzdata \
      curl && \
    ffmpeg -hide_banner -encoders 2>/dev/null | grep -qE 'h264_(qsv|vaapi|nvenc|amf)' || \
    (echo "WARN: no h264 HW encoders in ffmpeg build — software transcode only" >&2)

ARG SOMRA_UID=1000
ARG SOMRA_GID=1000
RUN addgroup -S -g ${SOMRA_GID} somra && \
    adduser -S -D -H -u ${SOMRA_UID} -G somra somra

RUN mkdir -p /data /config /media /cache /web/static && \
    chown -R somra:somra /data /config /cache /web

COPY --from=go-builder /build/bin/somra /usr/local/bin/somra
COPY --from=web-builder /build/web/dist/ /web/static/

ENV SOMRA_HTTP_ADDR=":8080" \
    SOMRA_DATA_DIR=/data \
    SOMRA_WEB_DIR=/web/static \
    SOMRA_LOG_LEVEL=info \
    SOMRA_LOG_FORMAT=json \
    PATH="/usr/local/bin:${PATH}"

USER somra:somra
WORKDIR /data

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -fsS --max-time 4 http://127.0.0.1:8080/api/v1/health || exit 1

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/somra"]
