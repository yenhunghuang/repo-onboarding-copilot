# Multi-stage build for secure and optimized container
ARG GO_VERSION=1.23
ARG ALPINE_VERSION=3.18

# Build stage
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

# Install security updates and build dependencies
RUN apk update && apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

# Create non-root user for building
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with security flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o repo-onboarding-copilot \
    ./cmd

# Final stage - minimal runtime image
FROM alpine:${ALPINE_VERSION}

# Install security updates and runtime dependencies
RUN apk update && apk add --no-cache \
    ca-certificates \
    tzdata \
    git \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Create app directory with proper permissions
RUN mkdir -p /app /tmp/analysis && \
    chown -R appuser:appgroup /app /tmp/analysis

# Copy binary from builder stage
COPY --from=builder --chown=appuser:appgroup /build/repo-onboarding-copilot /app/

# Copy configuration files if they exist
COPY --from=builder --chown=appuser:appgroup /build/configs/ /app/configs/ 2>/dev/null || true

# Set working directory
WORKDIR /app

# Switch to non-root user
USER appuser

# Expose port (if web interface is added later)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/app/repo-onboarding-copilot", "version"]

# Set entrypoint
ENTRYPOINT ["/app/repo-onboarding-copilot"]
CMD ["--help"]