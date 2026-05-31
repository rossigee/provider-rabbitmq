# Deployment Guide

## Image

```
ghcr.io/rossigee/provider-rabbitmq:<version>
```

## Crossplane package install

```bash
kubectl crossplane install provider ghcr.io/rossigee/provider-rabbitmq:<version>
```

## ProviderConfig

The `ProviderConfig` requires:

| Field | Description |
|-------|-------------|
| `spec.endpoint` | RabbitMQ Management API URL (`https://...`) — required |
| `spec.credentials.secretRef` | Secret containing `{"username":"...","password":"..."}` |
| `spec.tls.caBundleSecretRef` | Optional: secret key holding a PEM CA bundle for private CAs |

TLS verification is always enforced. The endpoint must use `https://`.

## Credentials secret format

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: rabbitmq-credentials
  namespace: crossplane-system
type: Opaque
stringData:
  credentials: '{"username":"admin","password":"secret"}'
```

## User passwords

User passwords must be stored in a separate Secret and referenced via
`spec.forProvider.passwordSecretRef` — never placed directly in the spec.

```yaml
spec:
  forProvider:
    name: app-user
    passwordSecretRef:
      namespace: default
      name: app-user-password
      key: password
```

## Monitoring

Apply `examples/monitoring.yaml` to create a `ServiceMonitor` and
`PrometheusRule` for the provider metrics endpoint (`:8080/metrics`).

Health probes are available at `:8080/healthz` and `:8080/readyz`.

## Building from source

```bash
go build -o provider ./cmd/provider
make docker-build
make xpkg-build
```

