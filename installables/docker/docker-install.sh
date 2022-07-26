#!/bin/bash
docker pull ghcr.io/middleware-labs/agent-host-go:dev
docker run -e TARGET=http://localhost:4317 -e MELT_API_KEY=avwisahcge0rx1hukcg4msbmlgcd2lijyx05 -d --name=MELT_Agent_go --restart always --network=host ghcr.io/middleware-labs/agent-host-go:dev
