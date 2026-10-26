# ── Stage 1: build ───────────────────────────────────────────────────────────
FROM golang:alpine AS builder

WORKDIR /app

# Copy dependency manifests first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

ENV GOARCH=arm64 GOOS=linux CGO_ENABLED=0

RUN go build -trimpath -ldflags="-s -w" -o chisa ./cmd

# Stage the bundled yt-dlp only when the checkout has one. Copying the whole
# directory always succeeds, unlike a COPY of bin/yt-dlp which breaks the build
# for anyone who never ran `make` (bin/ is gitignored). The runtime apk package
# is the fallback, so a missing bundle is not fatal.
# ponytail: no download step; upgrade path is a pinned `apk add yt-dlp=...`.
RUN mkdir -p /bundle && if [ -f bin/yt-dlp ]; then cp bin/yt-dlp /bundle/ && chmod +x /bundle/yt-dlp; fi

# ── Stage 2: runtime ──────────────────────────────────────────────────────────
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata ffmpeg yt-dlp

WORKDIR /app

COPY --from=builder /app/chisa .
COPY --from=builder /bundle/ ./bin/
COPY --from=builder /app/assets/ ./assets/

RUN chmod +x chisa && mkdir -p logs

CMD ["./chisa"]
