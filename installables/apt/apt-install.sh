#!/bin/bash
MW_LATEST_VERSION=0.0.14
export MW_LATEST_VERSION
export MW_AUTO_START=true

if [ "${MW_VERSION}" = "" ]; then 
  MW_VERSION=$MW_LATEST_VERSION
  export MW_VERSION
fi

# Adding APT repo address & public key to system
sudo mkdir -p /usr/local/bin/mw-go-agent/apt
sudo touch /usr/local/bin/mw-go-agent/apt/pgp-key-$MW_VERSION.public
sudo wget -O /usr/local/bin/mw-go-agent/apt/pgp-key-$MW_VERSION.public https://install.middleware.io/public-keys/pgp-key-$MW_VERSION.public
sudo touch /etc/apt/sources.list.d/mw-go.list

sudo mkdir -p /usr/bin/configyamls/all
sudo wget -O /usr/bin/configyamls/all/otel-config.yaml https://install.middleware.io/configyamls/all/otel-config.yaml
sudo mkdir -p /usr/bin/configyamls/metrics
sudo wget -O /usr/bin/configyamls/metrics/otel-config.yaml https://install.middleware.io/configyamls/metrics/otel-config.yaml
sudo mkdir -p /usr/bin/configyamls/traces
sudo wget -O /usr/bin/configyamls/traces/otel-config.yaml https://install.middleware.io/configyamls/traces/otel-config.yaml
sudo mkdir -p /usr/bin/configyamls/logs
sudo wget -O /usr/bin/configyamls/logs/otel-config.yaml https://install.middleware.io/configyamls/logs/otel-config.yaml
sudo mkdir -p /usr/bin/configyamls/nodocker
sudo wget -O /usr/bin/configyamls/nodocker/otel-config.yaml https://install.middleware.io/configyamls/nodocker/otel-config.yaml

echo "deb [arch=all signed-by=/usr/local/bin/mw-go-agent/apt/pgp-key-$MW_VERSION.public] https://install.middleware.io/repos/$MW_VERSION/apt-repo stable main" | sudo tee /etc/apt/sources.list.d/mw-go.list

# Updating apt list on system
sudo apt-get update -o Dir::Etc::sourcelist="sources.list.d/mw-go.list" -o Dir::Etc::sourceparts="-" -o APT::Get::List-Cleanup="0"

# Installing Agent
sudo apt-get install mw-go-agent-host

MW_USER=$(whoami)
export MW_USER

sudo su << EOSUDO


# Running Agent as a Daemon Service
touch /etc/systemd/system/mwservice.service

cat << EOF > /etc/systemd/system/mwservice.service
[Unit]
Description=Melt daemon!
[Service]
User=$MW_USER
#Code to execute
#Can be the path to an executable or code itself
WorkingDirectory=/usr/local/bin/mw-go-agent/apt
ExecStart=/usr/local/bin/mw-go-agent/apt/executable
Type=simple
TimeoutStopSec=10
Restart=on-failure
RestartSec=5
[Install]
WantedBy=multi-user.target
EOF

if [ ! "${TARGET}" = "" ]; then

cat << EOIF > /usr/local/bin/mw-go-agent/apt/executable
#!/bin/sh
cd /usr/bin && MW_API_KEY=$MW_API_KEY TARGET=$TARGET mw-go-agent-host start
EOIF

else 

cat << EOELSE > /usr/local/bin/mw-go-agent/apt/executable
#!/bin/sh
cd /usr/bin && MW_API_KEY=$MW_API_KEY mw-go-agent-host start
EOELSE

fi

chmod 777 /usr/local/bin/mw-go-agent/apt/executable

EOSUDO

sudo systemctl daemon-reload
sudo systemctl enable mwservice

if [ "${MW_AUTO_START}" = true ]; then	
    sudo systemctl start mwservice
fi


# Adding Cron to update + upgrade package every 5 minutes

sudo mkdir -p /usr/local/bin/mw-go-agent/apt/cron
sudo touch /usr/local/bin/mw-go-agent/apt/cron/mw-go.log

sudo crontab -l > cron_bkp
sudo echo "*/5 * * * * (wget -O /usr/local/bin/mw-go-agent/apt/pgp-key-$MW_VERSION.public https://install.middleware.io/public-keys/pgp-key-$MW_VERSION.public && sudo apt-get update -o Dir::Etc::sourcelist='sources.list.d/mw-go.list' -o Dir::Etc::sourceparts='-' -o APT::Get::List-Cleanup='0' && sudo apt-get install --only-upgrade telemetry-agent-host && sudo systemctl restart mwservice) >> /usr/local/bin/mw-go-agent/apt/cron/melt.log 2>&1 >> /usr/local/bin/mw-go-agent/apt/cron/melt.log" >> cron_bkp
sudo crontab cron_bkp
sudo rm cron_bkp


sudo su << EOSUDO

echo '

  Melt Go Agent Installed Successfully ! Happy MELTing !!
  ----------------------------------------------------

  /usr/local/bin 
    └───mw-go-agent
            └───apt: Contains all the required components to run APT package on the system
                └───executable: Contains the script to run agent
                └───pgp-key-$MW_VERSION.public: Contains copy of public key
                └───cron:
                    └───mw-go.log: Contains copy of public key

  /etc 
    ├─── apt
    |      └───sources.list.d
    |                └─── mw-go.list: Contains the APT repo entry
    └─── systemd
           └───system
                └─── mwservice.service: Service Entry for MW Agent
'
EOSUDO
