#!/bin/bash

echo "Checking service status..."
systemctl is-active --quiet iload-obd2 && echo "Service is running" || echo "Service is not running"

echo "Checking CAN interface..."
ip link show can0

echo "Checking logs..."
tail -n 20 /opt/iload-obd2/logs/iload-obd2.log

echo "Checking database access..."
sqlite3 /opt/iload-obd2/data/sqlite/vehicles.db ".tables"

echo "Checking web interface..."
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/