#!/bin/bash
# install.sh
# This script sets up iLoad-OBD2 to auto-start when CAN bus (can0) is detected on a Raspberry Pi

set -e

# Configuration
BASE_DIR="/opt/iload-obd2"
APP_PATH="$BASE_DIR/bin/iload-obd2"
CONFIG_PATH="$BASE_DIR/config/config.yaml"
DATA_DIR="$BASE_DIR/data"
LOG_DIR="$BASE_DIR/logs"
SERVICE_FILE="/etc/systemd/system/iload-obd2.service"
UDEV_RULE="/etc/udev/rules.d/99-canbus.rules"
LOGROTATE_FILE="/etc/logrotate.d/iload-obd2"
USER="$USER"
WORKDIR="$BASE_DIR"

# Check if service is running and stop it
if systemctl is-active --quiet iload-obd2; then
    echo "Stopping existing iLoad-OBD2 service..."
    sudo systemctl stop iload-obd2
fi

# Create directory structure (with -p to avoid errors if exists)
echo "Creating/verifying directory structure..."
sudo mkdir -p "$BASE_DIR"/{bin,config,logs}
sudo mkdir -p "$DATA_DIR"/{sqlite,influxdb}

# Backup existing config if present
if [ -f "$CONFIG_PATH" ]; then
    echo "Backing up existing config..."
    sudo cp "$CONFIG_PATH" "${CONFIG_PATH}.backup-$(date +%Y%m%d-%H%M%S)"
fi

# Set permissions (safe to run multiple times)
echo "Setting permissions..."
sudo chown -R "$USER:$USER" "$BASE_DIR"
sudo chmod -R 755 "$BASE_DIR"
sudo chmod -R 777 "$DATA_DIR"

# Create or update systemd service
echo "Updating systemd service..."
sudo tee "$SERVICE_FILE" > /dev/null <<EOF
[Unit]
Description=iLoad-OBD2 Telemetry Service
After=network.target influxdb.service
StartLimitIntervalSec=0

[Service]
Type=simple
User=$USER
ExecStart=$APP_PATH --config $CONFIG_PATH
WorkingDirectory=$WORKDIR
Environment=INFLUXDB_HOST=localhost
Environment=SQLITE_PATH=$DATA_DIR/sqlite/vehicles.db
Environment=LOG_PATH=$LOG_DIR/iload-obd2.log
Restart=always
RestartSec=1

[Install]
WantedBy=multi-user.target
EOF

# Set up log rotation (safe to overwrite)
echo "Updating log rotation config..."
sudo tee "$LOGROTATE_FILE" > /dev/null <<EOF
$LOG_DIR/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 644 $USER $USER
}
EOF

# Update udev rule
echo "Updating udev rule..."
sudo tee "$UDEV_RULE" > /dev/null <<EOF
SUBSYSTEM=="net", ACTION=="add", KERNEL=="can0", RUN+="/bin/systemctl start iload-obd2.service"
EOF

# Update CAN interface configuration (backup existing if present)
CAN_CONFIG="/etc/network/interfaces.d/can0"
if [ -f "$CAN_CONFIG" ]; then
    echo "Backing up existing CAN config..."
    sudo cp "$CAN_CONFIG" "${CAN_CONFIG}.backup-$(date +%Y%m%d-%H%M%S)"
fi

echo "Updating CAN interface configuration..."
sudo tee "$CAN_CONFIG" > /dev/null <<EOF
auto can0
iface can0 inet manual
    pre-up /sbin/ip link set \$IFACE type can bitrate 500000
    up /sbin/ip link set \$IFACE up
    down /sbin/ip link set \$IFACE down
EOF

# Reload configurations
echo "Reloading system configurations..."
sudo udevadm control --reload-rules
sudo systemctl daemon-reload

# Try to start the service if can0 is available
if ip link show can0 >/dev/null 2>&1; then
    echo "CAN interface found, starting service..."
    sudo systemctl start iload-obd2
else
    echo "CAN interface not found, service will start when can0 appears"
fi

# Verify installation
echo "Verifying installation..."
if systemctl is-active --quiet iload-obd2; then
    echo "Service is running"
else
    echo "Service is not running (waiting for CAN interface)"
fi

echo "Setup complete!"
echo "- Config backups (if any) are in: ${CONFIG_PATH}.backup-*"
echo "- Log files will be in: $LOG_DIR"
echo "- Data files will be in: $DATA_DIR"
echo "To enable the service at boot, run: sudo systemctl enable iload-obd2"