# mw-agent (Middleware Agent)

[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/middleware-labs/mw-agent)](https://goreportcard.com/report/github.com/middleware-labs/mw-agent)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![OpenTelemetry](https://img.shields.io/static/v1?label=Powered%20By&message=OpenTelemetry&labelColor=5c5c5c&color=1182c3&logo=opentelemetry&logoColor=white)

[![Linux](https://img.shields.io/badge/Linux-FCC624?logo=linux&logoColor=black)](#)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-326CE5?logo=kubernetes&logoColor=fff)](#)
[![Windows](https://custom-icon-badges.demolab.com/badge/Windows-0078D6?logo=windows11&logoColor=white)](#)
[![macOS](https://img.shields.io/badge/macOS-000000?logo=apple&logoColor=F0F0F0)](#)
[![Docker](https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=fff)](#)

## Overview

The Middleware Agent (`mw-agent`) collects and processes various types of monitoring data. It supports multiple deployment environments and provides comprehensive monitoring capabilities.

## Components

The agent consists of three main components:

1. **Host Agent** (`cmd/host-agent/`)
   - Designed for Linux, Windows & macOS environments
   - Collects system-level metrics and logs
   - Monitors Docker containers and applications

2. **Kubernetes Agent** (`cmd/kube-agent/`)
   - Specifically designed for Kubernetes environments
   - Collects cluster-level metrics and logs
   - Monitors pods, services, and other Kubernetes resources
   - Deployed as a daemonset & deployment

3. **Kubernetes Config Updater** (`cmd/kube-config-updater/`)
   - Manages configuration updates for Kubernetes deployments and daemonset
   - Ensures seamless configuration management in Kubernetes environments
   - Supports automated configuration updates

## Features

- **Multi-environment Support**: Deploy on Linux, Windows, or macOS
- **Comprehensive Monitoring**:
  - Metrics collection
  - Log aggregation
  - Trace collection
  - Synthetic monitoring
- **Flexible Configuration**: Support for environment variables, CLI flags, and configuration files
- **Docker Integration**: Native support for Docker container monitoring
- **Custom Tagging**: Ability to add custom tags to hosts and resources
- **Automated Configuration Management**: Seamless configuration updates in Kubernetes environments

## Installation

Check out various installation option available [here](https://docs.middleware.io/agent-installation/overview).

## Building from Source

The project includes a comprehensive Makefile for building different components:

```bash
# Build host agent for different platforms
make build-linux      # Linux
make build-windows    # Windows
make build-darwin-amd64  # macOS (Intel)
make build-darwin-arm64  # macOS (Apple Silicon)

# Build Kubernetes components
make build-kube              # Kubernetes agent
make build-kube-config-updater  # Kubernetes config updater

# Package components
make package-windows    # Windows installer
make package-linux-deb  # Debian package
make package-linux-rpm  # RPM package
make package-linux-docker  # Docker image
make package-kube-config-updater  # Kubernetes config updater image
```

## Configuration

The agent can be configured using:
- Environment variables
- Command-line flags
- Configuration file (YAML)

For detailed configuration options, see [Configuration Guide](docs/configuration.md).

### Basic Configuration Example

```yaml
api-key: YOUR_API_KEY
target: https://app.middleware.io
config-check-interval: 60s
docker-endpoint: unix:///var/run/docker.sock
host-tags: tag1=value1,tag2=value2
```

## Documentation

- [Installation Guide](https://docs.middleware.io/agent-installation/overview)
- [Configuration Guide](docs/configuration.md)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

