#!/bin/bash

# Stop the agent
sudo launchctl unload /Library/LaunchDaemons/io.middleware.mw-agent.plist

# Remove the LaunchDaemon plist file
sudo rm /Library/LaunchDaemons/io.middleware.mw-agent.plist

# Remove the agent files
sudo rm -rf /opt/mw-agent

# Remove the agent config files
sudo rm -rf /etc/mw-agent

# Remove the agent logs
sudo rm -rf /var/log/mw-agent

echo "Middleware Agent has been uninstalled."