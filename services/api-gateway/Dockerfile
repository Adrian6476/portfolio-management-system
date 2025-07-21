# Build stage
FROM golang:1.22.4-alpine AS builder

WORKDIR /app

# Install dependencies for building
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main .

# Final stage using distroless for maximum security
FROM gcr.io/distroless/static-debian12:nonroot

# Copy timezone data and ca-certificates from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from builder stage
COPY --from=builder /app/main /app/main

# Expose port
EXPOSE 8080

# Use non-root user for security
USER nonroot:nonroot

# Run the application
ENTRYPOINT ["/app/main"]