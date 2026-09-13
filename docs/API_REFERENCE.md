# RabbitMQ Provider API Reference

Complete API specification for provider-rabbitmq v1beta1 resources.

## API Groups

All RabbitMQ resources are organized by type under service-specific API groups:

- `vhost.rabbitmq.crossplane.io/v1beta1` — Virtual host management
- `exchange.rabbitmq.crossplane.io/v1beta1` — Exchange management
- `queue.rabbitmq.crossplane.io/v1beta1` — Queue management
- `binding.rabbitmq.crossplane.io/v1beta1` — Queue-to-exchange bindings
- `user.rabbitmq.crossplane.io/v1beta1` — User account management
- `rabbitmq.crossplane.io/v1beta1` — Provider configuration

## Common Fields

All resources share common Crossplane fields:

```yaml
apiVersion: <service>.rabbitmq.crossplane.io/v1beta1
kind: <ResourceType>
metadata:
  name: <resource-name>           # Kubernetes resource name
  namespace: <namespace>           # Namespace (required - namespaced resources)
spec:
  providerConfigRef:
    name: <provider-config-name>  # Reference to ProviderConfig
  forProvider:
    # Service-specific configuration
  deletionPolicy: Delete|Orphan   # How to handle deletion (default: Delete)
  managementPolicies:             # Control which operations are allowed
    - Create
    - Update
    - Delete
status:
  conditions: []                  # Readiness and error conditions
  atProvider:
    # Observed state from RabbitMQ
```

## ProviderConfig

Connection configuration for RabbitMQ provider.

```yaml
apiVersion: rabbitmq.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret                # Source of credentials (Secret required)
    secretRef:
      name: <secret-name>         # Secret name
      namespace: <namespace>       # Secret namespace
      key: credentials            # Key in secret containing JSON
```

**Credentials JSON Format:**
```json
{
  "username": "admin",
  "password": "password",
  "host": "rabbitmq.example.com",
  "port": 15672,
  "use_ssl": false
}
```

## Vhost Resource

Virtual host - logical separation of exchanges, queues, and users.

### Vhost Spec

```yaml
apiVersion: vhost.rabbitmq.crossplane.io/v1beta1
kind: Vhost
metadata:
  name: my-vhost
spec:
  forProvider:
    name: my_vhost                  # Virtual host name (required)
    description: "My VHost"         # Optional description
    tracing: false                  # Enable tracing (optional)
```

### Vhost Fields

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `name` | string | Virtual host identifier in RabbitMQ | Yes |
| `description` | string | Human-readable description | No |
| `tracing` | boolean | Enable message tracing for debugging | No |

### Vhost Status

```yaml
status:
  conditions:
    - type: Ready
      status: "True"
  atProvider:
    name: my_vhost
    messages_details: {...}
    messages: 0
    acks_details: {...}
```

## Exchange Resource

Message exchange - router for messages to queues based on routing rules.

### Exchange Spec

```yaml
apiVersion: exchange.rabbitmq.crossplane.io/v1beta1
kind: Exchange
metadata:
  name: my-exchange
spec:
  forProvider:
    name: my_exchange              # Exchange name (required)
    vhost: my_vhost                # Virtual host (required)
    type: topic                    # Exchange type: direct|topic|fanout|headers
    durable: true                  # Survives broker restart
    autoDelete: false              # Auto-delete when unused
    internal: false                # Internal only (not for direct publishing)
    arguments: {}                  # Additional broker arguments
```

### Exchange Types

| Type | Description | Use Case |
|------|-------------|----------|
| `direct` | Route to queues with exact routing key match | Point-to-point messaging |
| `topic` | Route to queues with pattern matching on routing key | Pub/sub with topics |
| `fanout` | Route to all bound queues | Broadcast messages |
| `headers` | Route based on message headers | Complex routing logic |

### Exchange Fields

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `name` | string | Exchange identifier | Yes |
| `vhost` | string | Virtual host containing exchange | Yes |
| `type` | string | Exchange type (direct, topic, fanout, headers) | Yes |
| `durable` | boolean | Persist across restarts | No (default: false) |
| `autoDelete` | boolean | Delete when last queue unbound | No (default: false) |
| `internal` | boolean | Restrict to internal message routing | No (default: false) |
| `arguments` | object | Broker-specific arguments | No |

## Queue Resource

Queue - storage for messages waiting for consumption.

### Queue Spec

```yaml
apiVersion: queue.rabbitmq.crossplane.io/v1beta1
kind: Queue
metadata:
  name: my-queue
spec:
  forProvider:
    name: my_queue                 # Queue name (required)
    vhost: my_vhost                # Virtual host (required)
    durable: true                  # Survives broker restart
    autoDelete: false              # Auto-delete when consumer disconnects
    exclusive: false               # Exclusive to connection (auto-delete)
    arguments:                      # Queue features
      x-message-ttl: 86400000      # Message TTL in milliseconds
      x-max-length: 10000          # Maximum queue length
      x-dead-letter-exchange: dlx  # Dead-letter exchange
```

### Queue Arguments

Common queue configuration:

| Argument | Type | Description |
|----------|------|-------------|
| `x-message-ttl` | integer (ms) | Message time-to-live |
| `x-max-length` | integer | Max messages in queue |
| `x-max-length-bytes` | integer | Max queue size in bytes |
| `x-dead-letter-exchange` | string | DLX for rejected messages |
| `x-dead-letter-routing-key` | string | DLK for dead-lettered messages |
| `x-expires` | integer (ms) | Queue auto-delete after inactivity |

### Queue Fields

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `name` | string | Queue identifier | Yes |
| `vhost` | string | Virtual host containing queue | Yes |
| `durable` | boolean | Persist across restarts | No (default: false) |
| `autoDelete` | boolean | Auto-delete on consumer disconnect | No (default: false) |
| `exclusive` | boolean | Exclusive to creator connection | No (default: false) |
| `arguments` | object | Queue configuration features | No |

## Binding Resource

Binding - connection between exchange and queue with routing rule.

### Binding Spec

```yaml
apiVersion: binding.rabbitmq.crossplane.io/v1beta1
kind: Binding
metadata:
  name: my-binding
spec:
  forProvider:
    vhost: my_vhost                # Virtual host (required)
    source: my_exchange            # Source exchange (required)
    destination: my_queue          # Destination queue/exchange (required)
    destinationType: queue         # "queue" or "exchange"
    routingKey: "order.*"          # Routing key pattern
    arguments: {}                  # Binding arguments
```

### Binding Fields

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `vhost` | string | Virtual host | Yes |
| `source` | string | Exchange name | Yes |
| `destination` | string | Queue or exchange name | Yes |
| `destinationType` | string | "queue" or "exchange" | Yes |
| `routingKey` | string | Routing key/pattern | Yes |
| `arguments` | object | Binding arguments | No |

### Routing Key Patterns

For `topic` exchanges:

- `*` — Match one word (e.g., "order.new" matches "order.*")
- `#` — Match zero or more words (e.g., "order.#" matches "order.new.pending")
- `order.new` — Exact match (works with direct exchanges)

## User Resource

User account for RabbitMQ access with permissions.

### User Spec

```yaml
apiVersion: user.rabbitmq.crossplane.io/v1beta1
kind: User
metadata:
  name: my-user
spec:
  forProvider:
    username: my_user              # Username (required)
    password: my_password          # Password (required)
    tags:                          # User tags/roles
      - administrator              # Options: administrator, management, policymaker
      - management
    limits: {}                      # Rate limiting
```

### User Tags

| Tag | Permissions |
|-----|-------------|
| `administrator` | Full admin access |
| `management` | Management API access |
| `policymaker` | Can set policies |
| (empty) | Standard user |

### User Fields

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `username` | string | Username identifier | Yes |
| `password` | string | User password | Yes |
| `tags` | array | User roles (administrator, management, policymaker) | No |
| `limits` | object | Rate limiting configuration | No |

## Status Fields

All resources report status through conditions and observations.

### Conditions

Standard Crossplane conditions:

| Type | Reason | Description |
|------|--------|-------------|
| `Ready` | `Available` | Resource created successfully |
| `Ready` | `Synced` | Resource in sync with RabbitMQ |
| `Ready` | `ReconcileFailed` | Sync failed |
| `Synced` | `ReconcileFailed` | Latest operation failed |

### Example Status

```yaml
status:
  conditions:
    - lastTransitionTime: "2025-09-13T10:00:00Z"
      message: "Resource is available"
      reason: "Available"
      status: "True"
      type: Ready
    - lastTransitionTime: "2025-09-13T10:00:00Z"
      message: ""
      reason: "ReconcileSuccess"
      status: "True"
      type: Synced
  atProvider:
    # Service-specific observed state
```

## Deletion Policies

Control deletion behavior:

| Policy | Behavior |
|--------|----------|
| `Delete` | Delete both CR and RabbitMQ resource (default) |
| `Orphan` | Delete only CR, keep RabbitMQ resource |

Use `Orphan` for critical resources you want to preserve:

```yaml
spec:
  deletionPolicy: Orphan
```

## Management Policies

Control which operations are allowed:

```yaml
spec:
  managementPolicies:
    - Create
    - Update
    - Delete
```

Options: `Create`, `Update`, `Delete`

Default: `[Create, Update, Delete]` (all operations allowed)

Example - read-only resource:

```yaml
spec:
  managementPolicies: []  # No operations allowed
```

## Examples

### Production Exchange with Dead-Letter Queue

```yaml
---
# Exchange
apiVersion: exchange.rabbitmq.crossplane.io/v1beta1
kind: Exchange
metadata:
  name: prod-orders
spec:
  forProvider:
    name: prod_orders
    vhost: production
    type: topic
    durable: true
---
# Dead-letter exchange
apiVersion: exchange.rabbitmq.crossplane.io/v1beta1
kind: Exchange
metadata:
  name: prod-dlx
spec:
  forProvider:
    name: prod_dlx
    vhost: production
    type: fanout
    durable: true
---
# Main queue with DLX
apiVersion: queue.rabbitmq.crossplane.io/v1beta1
kind: Queue
metadata:
  name: prod-orders-main
spec:
  forProvider:
    name: prod_orders_main
    vhost: production
    durable: true
    arguments:
      x-message-ttl: 3600000        # 1 hour TTL
      x-dead-letter-exchange: prod_dlx
---
# Dead-letter queue
apiVersion: queue.rabbitmq.crossplane.io/v1beta1
kind: Queue
metadata:
  name: prod-orders-dlq
spec:
  forProvider:
    name: prod_orders_dlq
    vhost: production
    durable: true
    arguments:
      x-message-ttl: 86400000       # 24 hour TTL for DLQ
---
# Bindings
apiVersion: binding.rabbitmq.crossplane.io/v1beta1
kind: Binding
metadata:
  name: prod-orders-binding
spec:
  forProvider:
    vhost: production
    source: prod_orders
    destination: prod_orders_main
    destinationType: queue
    routingKey: "order.#"
---
apiVersion: binding.rabbitmq.crossplane.io/v1beta1
kind: Binding
metadata:
  name: prod-dlq-binding
spec:
  forProvider:
    vhost: production
    source: prod_dlx
    destination: prod_orders_dlq
    destinationType: queue
    routingKey: ""
```

## API Compatibility

- **RabbitMQ Version**: 3.8+
- **Crossplane Version**: 1.14+
- **Kubernetes Version**: 1.20+

## Rate Limits

Check RabbitMQ documentation for management API rate limits when creating many resources programmatically.

## See Also

- [Getting Started Guide](getting-started.md)
- [RabbitMQ Documentation](https://www.rabbitmq.com/documentation.html)
- [Crossplane Documentation](https://docs.crossplane.io)
