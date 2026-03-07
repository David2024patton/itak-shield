# ── Build stage ──────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
COPY *.go ./
COPY audit/ audit/
COPY auth/ auth/
COPY cache/ cache/
COPY config/ config/
COPY dlp/ dlp/
COPY proxy/ proxy/
COPY retry/ retry/
COPY scanner/ scanner/
COPY server/ server/
COPY spend/ spend/
COPY tokenizer/ tokenizer/
COPY web/ web/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -X main.version=0.2.0" -o /itak-shield .

# ── Runtime stage ────────────────────────────
FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /itak-shield /usr/local/bin/itak-shield

# GUI port (default) and proxy port
EXPOSE 8080 8081

ENTRYPOINT ["itak-shield"]
# Default: GUI mode, accessible from the Docker host.
# Use: docker run -p 8080:8080 david2024patton/itak-shield
# CLI mode: docker run -p 8080:8080 david2024patton/itak-shield --target https://api.openai.com --port 8080 --bind 0.0.0.0
CMD ["--bind", "0.0.0.0", "--gui-port", "8080"]
