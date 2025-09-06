#!/bin/bash
# install-iload-obd2-autostart.sh
# This script sets up iLoad-OBD2 to auto-start when CAN bus (can0) is detected on a Raspberry Pi

set -e

APP_PATH="/usr/local/bin/iload-obd2"
CONFIG_PATH="/etc/iload-obd2/config.yaml"
SERVICE_FILE="/etc/systemd/system/iload-obd2.service"
UDEV_RULE="/etc/udev/rules.d/99-canbus.rules"
USER="pi"
WORKDIR="/home/pi"

# Create systemd service
sudo tee "$SERVICE_FILE" > /dev/null <<EOF
[Unit]
Description=iLoad-OBD2 Service
After=network.target

[Service]
ExecStart=$APP_PATH --config $CONFIG_PATH
Restart=on-failure
User=$USER
WorkingDirectory=$WORKDIR

[Install]
WantedBy=multi-user.target
EOF

# Create udev rule to start service when CAN bus appears
sudo tee "$UDEV_RULE" > /dev/null <<EOF
SUBSYSTEM=="net", ACTION=="add", KERNEL=="can0", RUN+="/bin/systemctl start iload-obd2.service"
EOF

# Reload udev and systemd
echo "Reloading udev and systemd..."
sudo udevadm control --reload-rules
sudo systemctl daemon-reload

echo "Setup complete. The service will start automatically when can0 appears."
echo "To enable the service at boot, run: sudo systemctl enable iload-obd2.service"
