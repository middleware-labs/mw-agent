# Debugging
We are using Delve to debug our code. 

## Building container

### For K8s
***NOTE:*** Before building image make sure you add break point to your code in VSCode

***Building a docker image***
```bash
docker build -t mw-agent-kube -f DockerfileKubeDebug .
```

***Running docker image***

```bash
docker run -e MW_API_KEY=<fetch_token_from_account> -e TARGET=<refer target list> -d --pid host --restart always --name mw-agent-k8-debug -p 4040:4040 docker.io/library/mw-agent-kube
```

### For Linux Agent
***NOTE:*** Before building image make sure you add break point to your code in VSCode

***Building a docker image***
```bash
docker build -t mw-agent-linux -f DockerfileLinuxDebug .
```

***Running docker image***

```bash
docker run -e MW_API_KEY=<fetch_token_from_account> -e TARGET=<refer target list> -d --pid host --restart always --name mw-agent-linux-debug -p 4040:4040 docker.io/library/mw-agent-linux
```

