FROM golang:1.25.0-alpine3.22 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build -ldflags="-w -s" -o iload-obd2

FROM alpine:latest

# Install required packages (different for Windows/Linux)
RUN if [ "$TARGETOS" = "linux" ]; then \
        apk add --no-cache sqlite influxdb tzdata curl; \
    fi

# Create app user
RUN adduser -D -u 1000 app

WORKDIR /app
COPY --from=builder /app/iload-obd2 .
COPY --from=builder /app/static ./static
COPY --from=builder /app/config.yaml .

# Create data directories
RUN mkdir -p /data/sqlite /data/influxdb && \
    chown -R app:app /app /data

USER app

# Environment variables
ENV SQLITE_PATH=/data/sqlite/vehicles.db
ENV INFLUXDB_PATH=/data/influxdb

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

EXPOSE 8080
VOLUME ["/data"]

ENTRYPOINT ["./iload-obd2"]
