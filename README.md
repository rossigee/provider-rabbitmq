# Provider RabbitMQ

[![CI](https://img.shields.io/github/actions/workflow/status/rossigee/provider-rabbitmq/ci.yml?branch=master)][build]
[![Version](https://img.shields.io/github/v/release/rossigee/provider-rabbitmq)][releases]
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

[build]: https://github.com/rossigee/provider-rabbitmq/actions/workflows/ci.yml
[releases]: https://github.com/rossigee/provider-rabbitmq/releases

A Crossplane v2 provider for managing RabbitMQ resources via the Management HTTP API. All resources are namespace-scoped for multi-tenancy.

## Container Registry

- **Primary**: `ghcr.io/rossigee/provider-rabbitmq:v0.1.0`

## Overview

A Crossplane provider for managing RabbitMQ resources including virtual hosts, exchanges, queues, and user permissions.

## Features

- **VHost Management**: Virtual host lifecycle management
- **Exchange Management**: Exchange declaration and configuration
- **Queue Management**: Queue declaration and binding
- **User Management**: User accounts with permissions
- **Multi-tenancy**: Namespace-scoped resources for team isolation

## Getting Started

### Prerequisites

- Kubernetes cluster with Crossplane installed
- RabbitMQ instance with Management HTTP API enabled
- RabbitMQ admin credentials

### Installation

```bash
kubectl crossplane install provider ghcr.io/rossigee/provider-rabbitmq:v0.1.0
```

### Configuration

Create a secret with your RabbitMQ credentials:

```bash
kubectl create secret generic rabbitmq-credentials \
  --from-literal=credentials='{"username":"admin","password":"secret"}' \
  -n crossplane-system
```

Create the ProviderConfig:

```yaml
apiVersion: rabbitmq.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  endpoint: https://rabbitmq.example.com:15672
  credentials:
    source: Secret
    secretRef:
      name: rabbitmq-credentials
      namespace: crossplane-system
      key: credentials
```

## Usage

### Create a Virtual Host

```yaml
apiVersion: rabbitmq.crossplane.io/v1beta1
kind: VHost
metadata:
  name: my-vhost
  namespace: production
spec:
  forProvider:
    name: /my-vhost
  providerConfigRef:
    name: default
```

### Create an Exchange

```yaml
apiVersion: rabbitmq.crossplane.io/v1beta1
kind: Exchange
metadata:
  name: my-exchange
  namespace: production
spec:
  forProvider:
    name: my-exchange
    vhostRef:
      name: my-vhost
    type: direct
    durable: true
  providerConfigRef:
    name: default
```

## Resource Types

| Resource | API Version | Description |
|----------|-------------|-------------|
| VHost | `rabbitmq.crossplane.io/v1beta1` | Virtual host |
| Exchange | `rabbitmq.crossplane.io/v1beta1` | Exchange declaration |
| Queue | `rabbitmq.crossplane.io/v1beta1` | Queue declaration |
| Binding | `rabbitmq.crossplane.io/v1beta1` | Exchange-to-queue binding |
| User | `rabbitmq.crossplane.io/v1beta1` | User account |
| Permission | `rabbitmq.crossplane.io/v1beta1` | Per-user, per-vhost ACL |

## Development

```bash
# Build the provider
make build

# Run tests
make test

# Lint code
make lint

# Generate CRDs
make generate
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

provider-rabbitmq is under the Apache 2.0 license.