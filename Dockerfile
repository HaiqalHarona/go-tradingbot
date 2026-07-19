# Build Stage
FROM golang:alpine AS builder

WORKDIR /app

# Install SSL certificates and build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy dependency manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bot main.go

# Final Lightweight Production Image
FROM alpine:latest

WORKDIR /app

# Install SSL certificates for HTTPS API calls to Alpaca
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from build stage
COPY --from=builder /app/bot /app/bot

# Run the trading bot engine
ENTRYPOINT ["/app/bot"]
