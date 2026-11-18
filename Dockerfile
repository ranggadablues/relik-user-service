# ==========================
# 1. BUILD STAGE
# ==========================
FROM golang:1.24.7-alpine AS builder

# Install git (needed for go mod download if fetching via VCS)
RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod and go.sum first (cache layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build statically for target platform (defaults to linux/amd64)
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o user-service ./cmd

# ==========================
# 2. RUN STAGE
# ==========================
FROM alpine:3.22

# Install CA certificates (required for Google/Facebook OAuth and email TLS) and create non-root user
RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S app -G app

WORKDIR /app

# Copy binary from build stage
COPY --from=builder --chown=app:app /app/user-service ./user-service

USER app

ENV SERVICE_PORT=50051

# Expose user gRPC port
EXPOSE 50051

# Run binary
CMD ["./user-service"]
