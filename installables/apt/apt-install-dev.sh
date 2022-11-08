#!/bin/bash
MW_LATEST_VERSION=0.0.14
export MW_LATEST_VERSION
export MW_AUTO_START=true

if [ "${MW_VERSION}" = "" ]; then 
  MW_VERSION=$MW_LATEST_VERSION
  export MW_VERSION
fi


MW_LOG_PATHS="/var/log/**/*.log"

echo -e "\nThe host agent will monitor all '.log' files inside your /var/log directory recursively [/var/log/**/*.log]"
while true; do
    
    read -p "Do you want to monitor any more directories for logs ? [Y|n] : " yn
    case $yn in
        [Yy]* )
          MW_LOG_PATH_DIR=""
          
          while true; do
            read -p "    Enter the absolute directory path from where you want to collect logs [/var/logs] : " MW_LOG_PATH_DIR
            export MW_LOG_PATH_DIR
            if [[ $MW_LOG_PATH_DIR =~ ^/|(/[\w-]+)+$ ]]
            then 
              break
            else
              echo "Invalid file path, try again ..."
            fi
          done

          while true; do
            read -p "    Do you want to watch "$MW_LOG_PATH_DIR" directory recursively ? (also watch files in subfolders) [Y|n] : " MW_LOG_PATH_DIR_RECURSIVE
            case $MW_LOG_PATH_DIR_RECURSIVE in
                [Yy]* ) 
                    MW_LOG_PATH_DIR=$MW_LOG_PATH_DIR/**/*
                    break;;
                [Nn]* )       
                    MW_LOG_PATH_DIR=$MW_LOG_PATH_DIR/*
                    break;;
                * )
                  echo "Please answer y or n"
                  continue;;
            esac
          done
          
          MW_LOG_PATH_DIR_EXTENSION=".log"
            while true; do
            read -p "    By default the agent will monitor '.log' files, do you want to replace the target extension ? [y|N] : " MW_LOG_PATH_DIR_EXTENSION_FLAG
            case $MW_LOG_PATH_DIR_EXTENSION_FLAG in
                [Yy]* ) 
                    read -p "    Enter extension that you want to watch [Ex => .json] : " MW_LOG_PATH_DIR_EXTENSION
                    break;;
                [Nn]* )
                    break;; 
                * )
                  echo "Please answer y or n"
                  continue;;              
            esac
          done

          MW_LOG_PATH_DIR=$MW_LOG_PATH_DIR$MW_LOG_PATH_DIR_EXTENSION;
          if [[ -n $MW_LOG_PATHS ]]
            then MW_LOG_PATHS=$MW_LOG_PATHS", "
          fi
          MW_LOG_PATHS=$MW_LOG_PATHS$MW_LOG_PATH_DIR
          echo -e "\nOur agent will now be monitoring these files : "$MW_LOG_PATHS
          continue;;
        [Nn]* ) 
          echo -e "\n----------------------------------------------------------\n\nOkay, Continuing installation ....\n\n----------------------------------------------------------\n"
          break;;
        * ) 
          echo -e "\nPlease answer y or n."
          continue;;
    esac
done

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
cd /usr/bin && MW_API_KEY=$MW_API_KEY TARGET=$TARGET MW_LOG_PATHS=$MW_LOG_PATHS mw-go-agent-host start
EOIF
else 
cat << EOELSE > /usr/local/bin/mw-go-agent/apt/executable
#!/bin/sh
cd /usr/bin && MW_API_KEY=$MW_API_KEY MW_LOG_PATHS=$MW_LOG_PATHS mw-go-agent-host start
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
  MW Go Agent Installed Successfully !
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