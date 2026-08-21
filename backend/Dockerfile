# Stage 1: Build the Go binary
FROM golang:1.25.5-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy dependency files first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build from the new cmd/server entry point
RUN CGO_ENABLED=0 GOOS=linux go build -o my-go-app ./cmd/server/...

# Stage 2: Final lightweight image
FROM alpine:latest

# Run as non-root user
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /app/my-go-app .
USER app

EXPOSE 8090

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8090/healthz || exit 1

CMD ["./my-go-app"]