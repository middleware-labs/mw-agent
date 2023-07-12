# mw-agent (Middleware Agent)
[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/middleware-labs/mw-agent)](https://goreportcard.com/report/github.com/middleware-labs/mw-agent)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![reviewdog](https://github.com/middleware-labs/mw-agent/actions/workflows/reviewdog.yml/badge.svg)](https://github.com/middleware-labs/mw-agent/actions/workflows/reviewdog.yml)

This repository contains code for the Middleware Agent (`mw-agent`). 

This agent currently supports [Linux](cmd/host-agent/)  & [Kubernetes](cmd/kube-agent/) environment.

## Project Structure
```text
mw-agent
    ├───configyamls: Set of otel-config.yamls based on MW_COLLECTION_TYPE filter for host agent.
    ├───configyamls-k8s: Set of otel-config.yamls for Kubernetes agent.
    ├───installables: Contains files required to install agent on Client system.
    |       └───docker: Contains Docker deployable script.
    |       └───apt: Contains files required to build APT package (Used in workflow).
    └───DockerfileLinux: To create multistage docker image for Linux environment.
    └───DockerfileKube: To create multistage docker image for Kubernetes environment.
    └───pkg: Contains common packages used for Linux & Kubernetes agents.
         └───config: Configuration options for the `mw-agent`.
```

## Installation & Configuration

`mw-agent` can take configuration from environment variables, CLI flags or configuration file. Details of how to configure the agent can be found [here](docs).


### Data collection
https://www.notion.so/Kubernetes-mw-c8826e2b20ac4a48b91fd98066924a13
