# provider-rabbitmq TODO

## Goal
Create a Crossplane provider for RabbitMQ that manages resources on external RabbitMQ clusters using a custom ProviderConfig CRD (instead of RabbitmqCluster reference).

## Current Status
**Build fixed** - All compilation errors resolved. Build and vet pass cleanly.

## What Was Done

### API Types Created
- `apis/v1beta1/` - ProviderConfig, ProviderCredentials
- `apis/vhost/v1beta1/` - Vhost managed resource
- `apis/exchange/v1beta1/` - Exchange managed resource
- `apis/queue/v1beta1/` - Queue managed resource
- `apis/binding/v1beta1/` - Binding managed resource
- `apis/user/v1beta1/` - User managed resource
- `apis/permission/v1beta1/` - Permission managed resource

### Controllers Created
- `internal/controller/vhost/vhost.go`
- `internal/controller/exchange/exchange.go`
- `internal/controller/queue/queue.go`
- `internal/controller/binding/binding.go`
- `internal/controller/user/user.go`
- `internal/controller/permission/permission.go`

### Client Created
- `internal/clients/rabbitmq.go` - Full RabbitMQ Management API client

### Build Fixes Applied
- Added manual `DeepCopyInto` methods for all Parameters/Observation/Spec/Status types
  (controller-gen v0.20.0 generates incorrect deepcopy code for this crossplane v2 pattern)
- Added `zz_generated.managed.go` files for all resource types implementing `resource.Managed`
  (`GetCondition`, `SetConditions`, `GetManagementPolicies`, `SetManagementPolicies`,
  `SetProviderConfigReference`, `GetWriteConnectionSecretToReference`, `SetWriteConnectionSecretToReference`)
- Fixed `feature.Feature` → `feature.Flag` in `internal/features/features.go`
- Fixed `internal/health/health.go` to use correct `healthz.Checker` signature (`func(*http.Request) error`)
- Fixed import aliases in `internal/clients/rabbitmq.go` (multiple v1beta1 packages need aliases)
- Fixed `GetConfig` to use `resource.CommonCredentialExtractor` instead of broken kube.Get approach
- Fixed duplicate import in `cmd/provider/main.go`
- Fixed unused variable/import warnings in controllers
- Removed leftover mailgun artefacts

## Next Steps

### Testing
```bash
cd /home/rossg/src/crossplane-providers/provider-rabbitmq

# Run unit tests
go test ./...

# Build local provider image
make local-provider

# Deploy to cluster
make up PROFILE=production
```

### Known Limitations
- `GetConfig` in `internal/clients/rabbitmq.go` assumes credentials secret contains JSON with `username`/`password` keys
- The vhost controller calls both `resource.CommonCredentialExtractor` (discarding result) and `clients.GetConfig`
  which also calls it - this is redundant but harmless. Consider refactoring to pass credentials directly.
- `InsecureSkipVerify: true` is hardcoded in the client - should be driven by `Config.Insecure`

## Files Reference
- Source: `/home/rossg/src/crossplane-providers/provider-rabbitmq/`
- Template: `/home/rossg/src/crossplane-providers/provider-rabbitmq/`
- Fixtures: `/home/rossg/infrastructure/flux-crossplane-rabbitmq/fixtures/`
