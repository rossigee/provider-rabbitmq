# Provider RabbitMQ

A Crossplane v2 provider for managing RabbitMQ resources via the Management HTTP API.
All resources are namespace-scoped for multi-tenancy.

## Supported Resources

| Kind | API Group | Description |
|------|-----------|-------------|
| `VHost` | `rabbitmq.crossplane.io/v1beta1` | Virtual host |
| `Exchange` | `rabbitmq.crossplane.io/v1beta1` | Exchange declaration |
| `Queue` | `rabbitmq.crossplane.io/v1beta1` | Queue declaration |
| `Binding` | `rabbitmq.crossplane.io/v1beta1` | Exchange-to-queue binding |
| `User` | `rabbitmq.crossplane.io/v1beta1` | User account |
| `Permission` | `rabbitmq.crossplane.io/v1beta1` | Per-user, per-vhost ACL |

## Quick Start

1. Create a credentials secret:
```bash
kubectl create secret generic rabbitmq-credentials \
  --from-literal=credentials='{"username":"admin","password":"secret"}' \
  -n crossplane-system
```

2. Apply the ProviderConfig:
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
      namespace: crossplane-system
      name: rabbitmq-credentials
      key: credentials
```

3. Apply sample resources:
```bash
kubectl apply -f examples/sample-resources.yaml
```

## TLS

TLS verification is always enforced. To use a private CA (e.g. managed by
trust-manager), set `spec.tls.caBundleSecretRef` on the ProviderConfig:

```yaml
spec:
  tls:
    caBundleSecretRef:
      namespace: crossplane-system
      name: rabbitmq-ca-bundle
      key: ca.crt
```

## User Passwords

User passwords are **never stored in the CR spec**. Reference a Secret instead:

```yaml
spec:
  forProvider:
    name: app-user
    passwordSecretRef:
      namespace: default
      name: app-user-password
      key: password
```

## Development

```bash
go build ./...
go test ./...
make generate   # regenerate deepcopy and CRDs after API changes
```

## License

Apache License 2.0. See [LICENSE](LICENSE).
