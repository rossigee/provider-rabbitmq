# Provider RabbitMQ Documentation

A Crossplane v2 provider for managing RabbitMQ topology. All managed resources are namespaced `rabbitmq.m.crossplane.io/v1beta1` with multi-tenancy support.

## Resource Documentation

| Resource | API Group | Description |
|----------|-----------|-------------|
| Vhost | `rabbitmq.m.crossplane.io/v1beta1` | Virtual hosts |
| Exchange | `rabbitmq.m.crossplane.io/v1beta1` | Exchanges |
| Queue | `rabbitmq.m.crossplane.io/v1beta1` | Queues |
| Binding | `rabbitmq.m.crossplane.io/v1beta1` | Exchange-to-queue bindings |
| User | `rabbitmq.m.crossplane.io/v1beta1` | Users |
| Permission | `rabbitmq.m.crossplane.io/v1beta1` | Per-user, per-vhost ACLs |
| ProviderConfig | `rabbitmq.m.crossplane.io/v1beta1` | Credentials (cluster-scoped) |

See `examples/sample-resources.yaml` for a complete topology example.

## API Coverage Gaps

RabbitMQ HTTP API surface not yet modeled: policies (HA/mirroring, TTL limits), operator policies, federation/shovel links, topic permissions beyond classic ACLs, global parameters, feature flags, and user password rotation without replacement.
