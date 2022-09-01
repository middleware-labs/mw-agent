#!/bin/bash
sudo systemctl stop meltgoservice
sudo systemctl disable meltgoservice

sudo apt-get remove mw-go-agent-host -y

sudo rm -rf /usr/local/bin/mw-go-agent/apt
sudo rm -rf /etc/systemd/system/meltgoservice.service
sudo rm -rf /etc/apt/sources.list.d/mw-go.list
sudo rm -rf /var/lib/apt/lists/host.middleware.io*
sudo apt-get clean
sudo apt autoremove
sudo crontab -r
sudo apt-get update