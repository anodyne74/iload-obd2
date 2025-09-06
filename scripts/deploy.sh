#!/bin/bash

# Configuration
APP_NAME="iload-obd2"
APP_USER="pi"
INSTALL_DIR="/opt/$APP_NAME"
DATA_DIR="/var/lib/$APP_NAME"
CONFIG_DIR="/etc/$APP_NAME"
LOG_DIR="/var/log/$APP_NAME"

# Build for Raspberry Pi (ARM)
echo "Building for Raspberry Pi..."
GOOS=linux GOARCH=arm64 go build -o $APP_NAME

# Create directories
echo "Creating directories..."
sudo mkdir -p $INSTALL_DIR
sudo mkdir -p $DATA_DIR
sudo mkdir -p $CONFIG_DIR
sudo mkdir -p $LOG_DIR

# Copy files
echo "Copying files..."
sudo cp $APP_NAME $INSTALL_DIR/
sudo cp -r static $INSTALL_DIR/
sudo cp config.yaml $CONFIG_DIR/
sudo cp scripts/iload-obd2.service /etc/systemd/system/

# Set permissions
echo "Setting permissions..."
sudo chown -R $APP_USER:$APP_USER $INSTALL_DIR
sudo chown -R $APP_USER:$APP_USER $DATA_DIR
sudo chown -R $APP_USER:$APP_USER $CONFIG_DIR
sudo chown -R $APP_USER:$APP_USER $LOG_DIR

# Create symlinks
sudo ln -sf $CONFIG_DIR/config.yaml $INSTALL_DIR/config.yaml

# Reload systemd
echo "Reloading systemd..."
sudo systemctl daemon-reload

# Enable and start service
echo "Enabling and starting service..."
sudo systemctl enable iload-obd2
sudo systemctl start iload-obd2

echo "Deployment complete!"
