# ── Build stage ────────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Cache module downloads independently of source changes
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# VERSION and BUILD_TIME can be overridden at build time (e.g. by CI).
# Example: docker build --build-arg VERSION=$(git rev-parse --short HEAD) ...
ARG VERSION=dev
ARG BUILD_TIME=""

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
      -ldflags="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}" \
      -o server \
      ./cmd/server

# ── Run stage ──────────────────────────────────────────────────────────────────
FROM alpine:3.21

# ca-certificates: required for TLS connections to Postgres
# tzdata: required for America/Chicago timezone used by the app
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/server ./server

# Templates and static files are loaded from disk at startup.
# They must live at web/templates/ and web/static/ relative to WORKDIR.
COPY web/ ./web/

EXPOSE 8080

CMD ["./server"]
