# ── Build stage ──────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod ./
COPY *.go ./
COPY proxy/ proxy/
COPY scanner/ scanner/
COPY server/ server/
COPY tokenizer/ tokenizer/
COPY web/ web/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -X main.version=0.1.0" -o /itak-shield .

# ── Runtime stage ────────────────────────────
FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /itak-shield /usr/local/bin/itak-shield

EXPOSE 8080

ENTRYPOINT ["itak-shield"]
CMD ["--target", "https://api.openai.com", "--port", "8080"]
