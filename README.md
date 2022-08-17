# Golang Host Agent

## Project Structure
```text
agent-host-go
    ├───installables: Contains files required to install agent on Client system
    |       └───docker: Contains Docker deployable script
    |       └───apt: Contains files required to build APT package (Used in workflow)
    └───Dockerfile: To create multistage docker image for Golang package
```

## Docker Installation
```
docker run -e MELT_API_KEY=<fetch_token_from_account> -e TARGET=<refer target list> -d --pid host --restart always ghcr.io/middleware-labs/agent-host-go:dev
```
OR create a `docker-compose.yml` with content given below :
```
version: "3.4"
services:  
  melt-agent-host-go:
    image: ghcr.io/middleware-labs/agent-host-go:dev
    restart: always
    pid: host
```
```
docker-compose up -d
```

## APT Installation

```
MELT_API_KEY=<fetch_token_from_your_account> TARGET=<refer target list> bash -c "$(curl -L https://host-go.melt.so/apt-install.sh)"
```
____________________________________________

### Advanced Options 


| ENV variables         | Usage            
| -------------         | ------------- 
| MELT_COLLECTION_TYPE          | Select among `metrics`, `traces` & `logs` to enable only a single pipeline
____________________________________________

### Target List

List available at https://github.com/middleware-labs/agent-host-rs README.md

----------------------------------------------

### Data collection
https://www.notion.so/Kubernetes-MELT-c8826e2b20ac4a48b91fd98066924a13
