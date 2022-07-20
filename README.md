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
docker run -e MELT_API_KEY=<fetch_token_from_account> -e TARGET=<refer target list> -d --pid host --restart always -v /var/run/docker.sock:/var/run/docker.sock ghcr.io/middleware-labs/agent-host:dev
```
OR create a `docker-compose.yml` with content given below :
```
version: "3.4"
services:  
  melt-agent-host-go:
    image: ghcr.io/middleware-labs/agent-host-go:dev
    restart: always
    pid: host
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
```
```
docker-compose up -d
```

## APT Installation

```
MELT_API_KEY=<fetch_token_from_your_account> TARGET=<refer target list> bash -c "$(curl -L https://host-go.melt.so/apt-install.sh)"
```
____________________________________________

### Target List

| Platform      | Target        |    
| ------------- | ------------- | 
| local         |  http://localhost:4317      |
| agent         |  https://us1-v1-grpc-agent.melt.so:9443   |
| front         |  https://us1-v1-grpc-front.melt.so:9443   |
| conflux       |  https://us1-v1-grpc-conflux.melt.so:9443 |
| capture       |  https://us1-v1-grpc-capture.melt.so:9443 |
| stage         |  https://us1-v1-grpc-stage.melt.so        |
| live          |  https://us1-v1-grpc.melt.so              |

*No need to specify target in live environment

----------------------------------------------

### Data collection
https://www.notion.so/Kubernetes-MELT-c8826e2b20ac4a48b91fd98066924a13
