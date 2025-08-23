FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev linux-headers

WORKDIR /app

# Copy only go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build with security flags
RUN CGO_ENABLED=1 GOOS=linux go build -buildvcs=false \
    -ldflags="-w -s" \
    -o iload-obd2

FROM alpine:latest

# Add non-root user
RUN addgroup -S app && adduser -S app -G app

# Install runtime dependencies
RUN apk add --no-cache \
    sqlite \
    influxdb \
    tzdata \
    can-utils \
    curl

WORKDIR /app
COPY --from=builder /app/iload-obd2 .
COPY --from=builder /app/static ./static
COPY --from=builder /app/config.yaml /etc/iload-obd2/config.yaml

# Set correct permissions
RUN chown -R app:app /app

# Create directories for data and logs
RUN mkdir -p /data/sqlite /data/influxdb /var/log/iload-obd2

# Set environment variables
ENV SQLITE_PATH=/data/sqlite/vehicles.db
ENV INFLUXDB_PATH=/data/influxdb
ENV CONFIG_PATH=/etc/iload-obd2/config.yaml

# Container metadata
LABEL org.opencontainers.image.source="https://github.com/anodyne74/iload-obd2" \
      org.opencontainers.image.description="iload-obd2 - OBD2 data collection and visualization" \
      org.opencontainers.image.licenses="MIT"

EXPOSE 8080
VOLUME ["/data", "/var/log/iload-obd2"]

# Switch to non-root user
USER app

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

CMD ["./iload-obd2", "--config", "/etc/iload-obd2/config.yaml"]
