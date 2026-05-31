# Crossplane Provider RabbitMQ Helm Chart

This Helm chart installs the Crossplane Provider for RabbitMQ in a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.20+
- Helm 3.0+
- Crossplane installed in the cluster

## Installation

```bash
helm install provider-rabbitmq ./charts/provider-rabbitmq \
  --namespace crossplane-system \
  --create-namespace
```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Provider image repository | `crossplane/provider-rabbitmq` |
| `image.tag` | Provider image tag | `v0.1.0` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `512Mi` |
| `resources.requests.cpu` | CPU request | `100m` |
| `resources.requests.memory` | Memory request | `128Mi` |
| `metrics.serviceMonitor.enabled` | Enable ServiceMonitor for Prometheus | `false` |
| `rbac.create` | Create RBAC resources | `true` |
| `podDisruptionBudget.enabled` | Enable PodDisruptionBudget | `false` |
| `networkPolicy.enabled` | Enable NetworkPolicy | `false` |

## Usage

Create a credentials secret and ProviderConfig — see the top-level
[README](../../README.md) and [examples/](../../examples/) for details.

## Troubleshooting

```bash
kubectl get providerconfigs
kubectl logs -n crossplane-system -l app.kubernetes.io/name=provider-rabbitmq
```

## License

Apache License 2.0 — see [LICENSE](../../LICENSE).
