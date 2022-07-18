#!/bin/bash
sudo systemctl stop meltgoservice
sudo systemctl disable meltgoservice

sudo apt-get remove melt-go-agent-host -y

sudo rm -rf /usr/local/bin/melt-go-agent/apt
sudo rm -rf /etc/systemd/system/meltgoservice.service
sudo rm -rf /etc/apt/sources.list.d/melt-go.list
sudo rm -rf /var/lib/apt/lists/host-go.melt.so*
sudo apt-get clean
sudo apt autoremove
sudo crontab -r
sudo apt-get update
# sudo rm -rf /var/lib/apt/lists/apt.melt.so*
# sudo apt-get update