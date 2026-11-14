# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git

# Shared proto contracts (go.mod replace => ../proto)
COPY proto ./proto

# Download dependencies
COPY url-service/go.mod url-service/go.sum ./url-service/
WORKDIR /src/url-service
RUN go mod download

# Copy source code
COPY url-service/ .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/url-service ./cmd/

# Production stage
FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/url-service .

EXPOSE 50051

CMD ["./url-service"]
