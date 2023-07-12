## Docker Installation
```
docker run -e MW_API_KEY=<fetch_token_from_account> -e TARGET=<refer target list> -d --pid host --restart always ghcr.io/middleware-labs/agent-host-go:dev
```
OR create a `docker-compose.yml` with content given below :
```
version: "3.4"
services:  
  mw-agent-host-go:
    image: ghcr.io/middleware-labs/agent-host-go:dev
    restart: always
    pid: host
```
```
docker-compose up -d
```

## APT Installation

```
MW_API_KEY=<fetch_token_from_your_account> TARGET=<refer target list> bash -c "$(curl -L https://host-go.melt.so/apt-install.sh)"
```
____________________________________________

### Advanced Options 


| ENV variables         | Usage            
| -------------         | ------------- 
| MW_COLLECTION_TYPE          | Select among `metrics`, `traces` & `logs` to enable only a single pipeline
____________________________________________

### Target List

List available at https://github.com/middleware-labs/agent-host-rs README.md