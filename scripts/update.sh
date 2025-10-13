#!/bin/bash
set -e

# Stop service if running
sudo systemctl stop iload-obd2 || true

# Install binary
sudo mv iload-obd2-arm64 /opt/iload-obd2/bin/iload-obd2
sudo chmod +x /opt/iload-obd2/bin/iload-obd2

# Update configuration
cp config.yaml /opt/iload-obd2/config/config.yaml

# Set permissions
sudo chown -R pi:pi /opt/iload-obd2
sudo chmod -R 755 /opt/iload-obd2
sudo chmod -R 777 /opt/iload-obd2/data

# Reload systemd and start service
sudo systemctl daemon-reload
sudo systemctl enable iload-obd2
sudo systemctl start iload-obd2

# Show status
sudo systemctl status iload-obd2