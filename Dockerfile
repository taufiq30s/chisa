# ── Stage 1: build ───────────────────────────────────────────────────────────
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Copy dependency manifests first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# buildx fills TARGETOS/TARGETARCH from `platforms:`; a plain `docker build`
# leaves them empty and Go falls back to the host. Never hardcode these: a
# hardcoded arch makes the image unrunnable on every other machine.
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o chisa ./cmd

# ── Stage 2: yt-dlp ──────────────────────────────────────────────────────────
# Upstream's musl build is a self-contained binary, so the runtime image needs
# no python3 (~70 MB). Pinning the APK package instead would freeze an extractor
# that YouTube breaks within weeks, and yt-dlp advises against self-updating
# packaged installs. `latest/download` tracks releases, SHA2-256SUMS keeps the
# download honest.
FROM alpine:3.24 AS ytdlp

RUN apk add --no-cache ca-certificates curl

ARG TARGETARCH
RUN set -eu; \
    case "${TARGETARCH:-amd64}" in \
        amd64) asset=yt-dlp_musllinux ;; \
        arm64) asset=yt-dlp_musllinux_aarch64 ;; \
        *) echo "unsupported TARGETARCH: ${TARGETARCH}" >&2; exit 1 ;; \
    esac; \
    base=https://github.com/yt-dlp/yt-dlp/releases/latest/download; \
    curl -fsSL -o /yt-dlp "$base/$asset"; \
    curl -fsSL -o /sums "$base/SHA2-256SUMS"; \
    hash=$(awk -v f="$asset" '$2 == f { print $1 }' /sums); \
    [ -n "$hash" ] || { echo "no checksum published for $asset" >&2; exit 1; }; \
    echo "$hash  /yt-dlp" | sha256sum -c -; \
    chmod +x /yt-dlp; \
    rm /sums

# ── Stage 3: runtime ─────────────────────────────────────────────────────────
FROM alpine:3.24

RUN apk add --no-cache ca-certificates tzdata ffmpeg

WORKDIR /app

COPY --from=builder /app/chisa .
COPY --from=builder /app/assets/ ./assets/
# FindYtdlpBinary prefers ./bin/yt-dlp over PATH, so no code change is needed.
COPY --from=ytdlp /yt-dlp ./bin/yt-dlp

# yt-dlp scans <exe-dir>/yt-dlp-plugins/<name>/yt_dlp_plugins/ with no flag, so
# a host directory mounted here is picked up without a Go change or rebuild.
RUN chmod +x chisa bin/yt-dlp && mkdir -p logs bin/yt-dlp-plugins

CMD ["./chisa"]
