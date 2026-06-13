# Stage 1: Build the Go binary
FROM golang:1.25.5-alpine AS builder

# Install git because some Go modules require git to download
RUN apk add --no-cache git

WORKDIR /app

# Copy EVERYTHING first to make sure go.mod and all files are accessible
COPY . .

# Clean the mod cache and force Go to pull/tidy dependencies cleanly
RUN go mod tidy
RUN go mod download

# Build the application binary
RUN CGO_ENABLED=0 GOOS=linux go build -o my-go-app .

# Stage 2: Final lightweight image
FROM alpine:latest
WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/my-go-app .

EXPOSE 8090

CMD ["./my-go-app"]
