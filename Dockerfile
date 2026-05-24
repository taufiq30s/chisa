# ── Stage 1: build ───────────────────────────────────────────────────────────
FROM golang:alpine AS builder

WORKDIR /app

# Copy dependency manifests first for better layer caching
COPY go.mod go.sum ./
COPY .vendor/ .vendor/
RUN go mod download

# Copy source
COPY . .

# In CI, lavalink_nodes.json may be absent because it is gitignored.
# Fallback to the committed backup so runtime COPY does not fail.
RUN if [ ! -f /app/lavalink_nodes.json ] && [ -f /app/lavalink_nodes.json.backup ]; then \
			cp /app/lavalink_nodes.json.backup /app/lavalink_nodes.json; \
		fi

ENV GOARCH=arm64 GOOS=linux CGO_ENABLED=0

RUN go build -trimpath -ldflags="-s -w" -o chisa ./cmd

# ── Stage 2: runtime ──────────────────────────────────────────────────────────
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/chisa .
# COPY --from=builder /app/lavalink_nodes.json .
COPY --from=builder /app/assets/ ./assets/

RUN chmod +x chisa && mkdir -p logs

CMD ["./chisa"]
