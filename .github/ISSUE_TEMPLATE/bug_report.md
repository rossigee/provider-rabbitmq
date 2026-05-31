## Bug Report: ProviderConfigUsage API Group Mismatch

### Description
The provider-rabbitmq has an API group mismatch for ProviderConfigUsage resources. The provider code references `rabbitmq.crossplane.io/v1beta1` but the CRD and API are registered under `rabbitmq.m.crossplane.io/v1beta1`.

### Impact
- Provider cannot track ProviderConfig usage
- Resources fail to sync with error: `cannot get restmapping: no matches for kind "ProviderConfigUsage" in version "rabbitmq.crossplane.io/v1beta1"`
- Domain reconciliation fails with "CannotConnectToProvider" errors

### Reproduction Steps
1. Install provider-rabbitmq v0.14.3 on Crossplane
2. Create a Domain CR referencing a ProviderConfig
3. Observe provider logs showing ProviderConfigUsage errors

### Expected Behavior
Provider should be able to track ProviderConfig usage without errors.

### Root Cause
In `apis/v1beta1/register.go`:
```go
Group = "rabbitmq.m.crossplane.io"
```

But the provider code attempts to use `rabbitmq.crossplane.io` when creating ProviderConfigUsage resources.

### Suggested Fix
Update the provider to use the correct API group (`rabbitmq.m.crossplane.io`) when creating/updating ProviderConfigUsage resources, or ensure consistency across the codebase.

### Environment
- provider-rabbitmq: v0.14.3
- Crossplane: v2.3.0
- Kubernetes: v1.35.1 (minikube)

### Additional Notes
- Also observed RBAC issues - provider service account needs broader permissions including events and status updates
