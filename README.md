# mw-agent (Middleware Agent)
[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/middleware-labs/mw-agent)](https://goreportcard.com/report/github.com/middleware-labs/mw-agent)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![reviewdog](https://github.com/middleware-labs/mw-agent/actions/workflows/reviewdog.yml/badge.svg)](https://github.com/middleware-labs/mw-agent/actions/workflows/reviewdog.yml)

This repository contains code for the Middleware Agent (`mw-agent`). 

This agent currently supports [Linux & Windows](cmd/host-agent/) & [Kubernetes](cmd/kube-agent/) environment.


## Installation & Configuration

`mw-agent` can take configuration from environment variables, CLI flags or configuration file. Details of how to configure the agent can be found [here](docs).


----

https://test-keval.free.beeceptor.com/aws-jobs

 helm install --set mw.target=https://4plo493.middleware.io:443 --set mw.apiKey=evaddjfmazsdz8qip2cxva99muxv30wq6g6c --set clusterMetadata.name=minikube-keval --wait mw-aws-data-scraper -n mw-agent-ns --create-namespace 

helm upgrade --install --set mw.target=https://4plo493.middleware.io:443 --set mw.apiKey=evaddjfmazsdz8qip2cxva99muxv30wq6g6c --set clusterMetadata.name=minikube-keval --wait  my-mw-aws-data-scraper  mw-aws-data-scraper -n mw-agent-ns

  minikube addons enable metrics-server

  https://linear.app/middleware/issue/ENG-2933/create-a-server-side-component-to-do-aws-metrics-scrapping

  minikube image load ghcr.io/middleware-labs/mw-kube-agent:awscloudwatchmetricsreceiver-keval