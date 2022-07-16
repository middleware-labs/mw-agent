#!/bin/bash
sudo systemctl stop meltgoservice
sudo systemctl disable meltgoservice
sudo apt-get remove melt-go-agent-host -y
sudo rm -rf $HOME/melt-go-agent/apt
sudo apt-get clean
sudo apt autoremove
# sudo rm -rf /var/lib/apt/lists/apt.melt.so*
# sudo apt-get update