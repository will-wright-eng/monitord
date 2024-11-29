#!/bin/bash
set -e
# Build the binary
go build -o monitord ./cmd/monitord
# Create necessary directories
sudo mkdir -p /usr/local/bin
sudo mkdir -p /usr/local/var/log
# Install binary
sudo cp monitord /usr/local/bin/
sudo chmod +x /usr/local/bin/monitord
# Install LaunchAgent
cp com.username.monitord.plist ~/Library/LaunchAgents/
# Load the LaunchAgent
launchctl load ~/Library/LaunchAgents/com.username.monitord.plist
echo "Installation complete!"
