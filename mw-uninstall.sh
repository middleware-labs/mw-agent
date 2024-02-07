#!/bin/bash

# stopping and removing the service
sudo systemctl stop mwservice
sudo systemctl disable mwservice
sudo rm -rf /etc/systemd/system/mwservice.service

# deleting the MW agent binary
sudo apt-get purge mw-agent -y

# deleting MW agent artifacts
sudo rm -rf /usr/local/bin/mw-agent

# deleting entry from APT list
sudo rm -rf /etc/apt/sources.list.d/mw-agent.list
