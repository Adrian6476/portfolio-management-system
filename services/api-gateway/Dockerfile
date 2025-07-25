# Build stage
FROM crpi-uq4h40nqjrf2s1wo.cn-qingdao.personal.cr.aliyuncs.com/ghcr_io/golang:1.22.4-alpine AS builder

WORKDIR /app

# Install dependencies for building
# Use Chinese mirror for Alpine packages
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
# Use Chinese proxy for Go modules
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main .

# Final stage using distroless for maximum security
FROM crpi-uq4h40nqjrf2s1wo.cn-qingdao.personal.cr.aliyuncs.com/ghcr_io/static-debian12:nonroot

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