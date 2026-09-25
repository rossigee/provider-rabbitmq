# Getting Started with Provider RabbitMQ

Complete guide to installing and using provider-rabbitmq v1beta1 with Crossplane.

## Prerequisites

- **Kubernetes cluster** with Crossplane v2.5+ installed
- **RabbitMQ** instance with Management HTTP API enabled (usually on port 15672)
- RabbitMQ admin credentials
- `kubectl` configured to access your cluster

## Installation

### 1. Install Crossplane

```bash
helm repo add crossplane-stable https://charts.crossplane.io/stable
helm install crossplane \
  crossplane-stable/crossplane \
  -n crossplane-system \
  --create-namespace
```

Wait for Crossplane to be ready:
```bash
kubectl wait -n crossplane-system --for=condition=Ready pods -l app.kubernetes.io/instance=crossplane --timeout=300s
```

### 2. Install Provider RabbitMQ

```bash
# Using Crossplane CLI
kubectl crossplane install provider ghcr.io/rossigee/provider-rabbitmq:v0.5.3

# Or using Helm
helm repo add crossplane-contrib https://charts.crossplane.io/contrib
helm install provider-rabbitmq \
  crossplane-contrib/provider-rabbitmq \
  -n crossplane-system \
  --version ">=0.5.3"
```

Verify installation:
```bash
kubectl get providers
kubectl describe provider provider-rabbitmq
```

## Configuration

### 1. Create RabbitMQ Credentials Secret

Create a secret containing RabbitMQ Management API credentials:

```bash
kubectl create secret generic rabbitmq-creds \
  -n default \
  --from-literal=credentials='
{
  "username": "guest",
  "password": "guest",
  "host": "rabbitmq.example.com",
  "port": 15672,
  "use_ssl": false
}'
```

Or using a file:

```bash
cat > rabbitmq-creds.json <<EOF
{
  "username": "rabbitmq-admin",
  "password": "secure-password",
  "host": "rabbitmq.example.com",
  "port": 15672,
  "use_ssl": true
}
EOF

kubectl create secret generic rabbitmq-creds \
  -n default \
  --from-file=credentials=rabbitmq-creds.json
```

### 2. Create ProviderConfig

Apply the ProviderConfig to connect to RabbitMQ:

```yaml
apiVersion: rabbitmq.m.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
  namespace: default
spec:
  credentials:
    source: Secret
    secretRef:
      name: rabbitmq-creds
      namespace: default
      key: credentials
```

Save as `provider-config.yaml` and apply:
```bash
kubectl apply -f provider-config.yaml
```

Verify:
```bash
kubectl get providerconfig
kubectl describe providerconfig default
```

## First Resource: Create a Virtual Host

### 1. Create a Virtual Host

Virtual hosts provide logical separation in RabbitMQ:

```yaml
apiVersion: rabbitmq.m.crossplane.io/v1beta1
kind: Vhost
metadata:
  name: demo-vhost
  namespace: default
spec:
  providerConfigRef:
    name: default
  
  forProvider:
    name: demo_vhost
    description: "Demo virtual host"
    tracing: false
  
  deletionPolicy: Delete
```

Apply:
```bash
kubectl apply -f vhost.yaml
```

Monitor creation:
```bash
kubectl describe vhost demo-vhost -n default
kubectl get vhost demo-vhost -o jsonpath='{.status.conditions}'
```

## Common Tasks

### Create an Exchange

```yaml
apiVersion: rabbitmq.m.crossplane.io/v1beta1
kind: Exchange
metadata:
  name: demo-exchange
  namespace: default
spec:
  providerConfigRef:
    name: default
  
  forProvider:
    name: demo_exchange
    vhost: demo_vhost
    type: topic
    durable: true
    autoDelete: false
    internal: false
  
  deletionPolicy: Delete
```

### Create a Queue

```yaml
apiVersion: rabbitmq.m.crossplane.io/v1beta1
kind: Queue
metadata:
  name: demo-queue
  namespace: default
spec:
  providerConfigRef:
    name: default
  
  forProvider:
    name: demo_queue
    vhost: demo_vhost
    durable: true
    autoDelete: false
    exclusive: false
  
  deletionPolicy: Delete
```

### Create a Binding

Bind a queue to an exchange:

```yaml
apiVersion: rabbitmq.m.crossplane.io/v1beta1
kind: Binding
metadata:
  name: demo-binding
  namespace: default
spec:
  providerConfigRef:
    name: default
  
  forProvider:
    vhost: demo_vhost
    source: demo_exchange
    destination: demo_queue
    destinationType: queue
    routingKey: "demo.*"
  
  deletionPolicy: Delete
```

### Create a User

```yaml
apiVersion: rabbitmq.m.crossplane.io/v1beta1
kind: User
metadata:
  name: demo-user
  namespace: default
spec:
  providerConfigRef:
    name: default
  
  forProvider:
    username: demo_user
    password: demo_password
    tags:
      - administrator
  
  deletionPolicy: Delete
```

## Complete Example

Create a full messaging setup:

```yaml
---
# Virtual Host
apiVersion: rabbitmq.m.crossplane.io/v1beta1
kind: Vhost
metadata:
  name: messaging
  namespace: default
spec:
  providerConfigRef:
    name: default
  forProvider:
    name: messaging_vhost
    description: "Production messaging virtual host"
---
# Exchange for topic routing
apiVersion: rabbitmq.m.crossplane.io/v1beta1
kind: Exchange
metadata:
  name: topics
  namespace: default
spec:
  providerConfigRef:
    name: default
  forProvider:
    name: amq.topic
    vhost: messaging_vhost
    type: topic
    durable: true
---
# Queue for events
apiVersion: rabbitmq.m.crossplane.io/v1beta1
kind: Queue
metadata:
  name: events
  namespace: default
spec:
  providerConfigRef:
    name: default
  forProvider:
    name: events_queue
    vhost: messaging_vhost
    durable: true
---
# Bind queue to exchange
apiVersion: rabbitmq.m.crossplane.io/v1beta1
kind: Binding
metadata:
  name: events-binding
  namespace: default
spec:
  providerConfigRef:
    name: default
  forProvider:
    vhost: messaging_vhost
    source: amq.topic
    destination: events_queue
    destinationType: queue
    routingKey: "events.*"
---
# Application user
apiVersion: rabbitmq.m.crossplane.io/v1beta1
kind: User
metadata:
  name: app-user
  namespace: default
spec:
  providerConfigRef:
    name: default
  forProvider:
    username: app_user
    password: app_password
    tags: []
```

Apply entire stack:
```bash
kubectl apply -f complete-setup.yaml
```

## List Resources

```bash
# List all virtual hosts
kubectl get vhosts -n default

# List all exchanges
kubectl get exchanges -n default

# List all queues
kubectl get queues -n default

# List all bindings
kubectl get bindings -n default

# List all users
kubectl get users -n default
```

## Describe Resources

```bash
kubectl describe vhost messaging -n default
kubectl describe exchange topics -n default
kubectl describe queue events -n default
```

## Delete Resources

```bash
# Delete individual resource
kubectl delete vhost messaging -n default

# Delete entire setup
kubectl delete -f complete-setup.yaml
```

With `deletionPolicy: Delete` (default), both the Kubernetes resource and RabbitMQ resource are deleted. Use `deletionPolicy: Orphan` to keep the RabbitMQ resource when deleting the CR.

## Troubleshooting

### Provider Not Ready

```bash
# Check provider status
kubectl get provider provider-rabbitmq
kubectl describe provider provider-rabbitmq

# Check logs
kubectl logs -n crossplane-system -l pkg.crossplane.io/provider=provider-rabbitmq
```

### Resource Stuck in Creating

```bash
# Check resource status
kubectl describe vhost messaging -n default

# Check conditions
kubectl get vhost messaging -o jsonpath='{.status.conditions}' | jq

# Check provider logs
kubectl logs -n crossplane-system -l pkg.crossplane.io/provider=provider-rabbitmq -f
```

### Connection Errors

Verify credentials:
```bash
# Check secret exists
kubectl get secret rabbitmq-creds -n default

# Verify format
kubectl get secret rabbitmq-creds -n default -o jsonpath='{.data.credentials}' | base64 -d | jq

# Test connectivity
curl -u admin:guest https://rabbitmq.example.com:15672/api/vhosts
```

### Authentication Issues

- Verify username and password in credentials secret
- Confirm user has admin permissions in RabbitMQ
- Check Management HTTP API is enabled: `rabbitmq-plugins enable rabbitmq_management`

## Best Practices

### Namespace Isolation

Use separate namespaces for different environments:

```bash
kubectl create namespace production
kubectl apply -f provider-config.yaml -n production
kubectl apply -f messaging-resources.yaml -n production
```

### Deletion Policies

- `deletionPolicy: Delete` — Delete RabbitMQ resource when CR is deleted
- `deletionPolicy: Orphan` — Keep RabbitMQ resource, only delete CR

Choose based on your disaster recovery requirements.

### Resource Naming

Use descriptive names with environment prefix:
- `prod-events-queue`
- `staging-api-exchange`
- `dev-test-binding`

## Next Steps

1. **Explore all resources**: See resources section for complete API
2. **Advanced configuration**: Use management policies for fine-grained control
3. **Monitor resources**: Check status and conditions regularly
4. **Automate deployments**: Version control your manifests

## Getting Help

- **GitHub Issues**: https://github.com/rossigee/provider-rabbitmq/issues
- **Crossplane Slack**: https://slack.crossplane.io
- **RabbitMQ Docs**: https://www.rabbitmq.com/documentation.html
