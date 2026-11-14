# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates git

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/url-service ./cmd/

# Production stage
FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/url-service .

EXPOSE 8080
EXPOSE 50051

CMD ["./url-service"]
