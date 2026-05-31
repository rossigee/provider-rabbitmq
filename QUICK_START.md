# Quick Start

## Prerequisites

- Crossplane installed in your cluster
- A RabbitMQ instance with the Management API enabled (port 15671/15672)

## 1. Install the provider

```bash
kubectl crossplane install provider ghcr.io/rossigee/provider-rabbitmq:latest
kubectl wait --for=condition=Installed provider.pkg.crossplane.io/provider-rabbitmq --timeout=120s
```

## 2. Create a credentials secret

```bash
kubectl create secret generic rabbitmq-credentials \
  --from-literal=credentials='{"username":"admin","password":"your-password"}' \
  -n crossplane-system
```

## 3. Create a ProviderConfig

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

If your RabbitMQ uses a private CA (e.g. issued by trust-manager):

```yaml
spec:
  tls:
    caBundleSecretRef:
      namespace: crossplane-system
      name: rabbitmq-ca
      key: ca.crt
```

## 4. Create resources

```bash
kubectl apply -f examples/sample-resources.yaml
```

This creates a vhost, exchange, queue, binding, user, and permission.

## Verify

```bash
kubectl get vhosts,exchanges,queues,bindings,users,permissions
```

