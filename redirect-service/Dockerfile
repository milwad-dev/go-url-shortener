# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git

# Shared proto contracts (go.mod replace => ../proto)
COPY proto ./proto

# Download dependencies
COPY redirect-service/go.mod redirect-service/go.sum ./redirect-service/
WORKDIR /src/redirect-service
RUN go mod download

# Copy source code
COPY redirect-service/ .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/redirect-service ./cmd/

# Production stage
FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/redirect-service .

EXPOSE 8081

CMD ["./redirect-service"]
