# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Project Overview

This is a **Crossplane v2 provider** for managing RabbitMQ resources via the
RabbitMQ Management HTTP API. All managed resources are namespace-scoped.

**Supported resources** (all `v1beta1`, namespace-scoped):
- `VHost` — virtual host
- `Exchange` — exchange declaration
- `Queue` — queue declaration
- `Binding` — exchange-to-queue binding
- `User` — user account (password via `PasswordSecretRef`, never in spec)
- `Permission` — per-user, per-vhost ACL

**API group**: `rabbitmq.crossplane.io`

## Directory Structure

```
apis/
  v1beta1/              # ProviderConfig, ProviderConfigUsage
  binding/v1beta1/
  exchange/v1beta1/
  permission/v1beta1/
  queue/v1beta1/
  user/v1beta1/
  vhost/v1beta1/
internal/
  clients/rabbitmq.go   # RabbitMQ Management API client + Client interface
  controller/
    binding/
    exchange/
    permission/
    queue/
    user/
    vhost/
    controller.go       # controller registration
  features/features.go
  health/health.go
  version/version.go
cmd/provider/main.go
package/
  crds/                 # CRD YAML files (prefixed with _)
  crossplane.yaml       # package metadata
examples/
  provider/             # ProviderConfig and credentials secret template
  sample-resources.yaml # one example of each resource type
```

## Build Commands

```bash
go build ./...          # compile
go test ./...           # run tests
make generate           # regenerate deepcopy, managed resource methods, CRDs
make lint               # run golangci-lint
make docker-build       # build container image
make xpkg-build         # build Crossplane .xpkg package
```

## Controller Pattern

Every resource controller lives in `internal/controller/<resource>/` and
implements the `managed.ExternalClient` interface:

```go
func (c *external) Observe(ctx, mg) (ExternalObservation, error)
func (c *external) Create(ctx, mg)  (ExternalCreation, error)
func (c *external) Update(ctx, mg)  (ExternalUpdate, error)
func (c *external) Delete(ctx, mg)  (ExternalDelete, error)
```

Key rules:
- `Observe` must be **read-only** — never call mutating API methods from Observe.
- Use `clients.IsNotFound(err)` to distinguish missing resources from real errors.
- Set `xpv1.Available()` on the managed resource after successful Observe/Create.
- `Update` is a no-op for resources that cannot be changed after creation.

## Client Interface

`internal/clients/rabbitmq.go` defines the `Client` interface and the
`rabbitmqClient` struct. The `GetConfig` function builds a `*Config` from the
referenced `ProviderConfig` and resolves credentials from the referenced Secret.

TLS:
- Always enforced (`MinVersion: tls.VersionTLS12`).
- Custom CA supplied via `spec.tls.caBundleSecretRef` on the ProviderConfig.
- `AppendCertsFromPEM` return value is checked; bad PEM is surfaced as an error
  via `errClient` rather than silently falling back to the system pool.

HTTP:
- `request()` reads responses with `io.ReadAll(resp.Body)`.
- Error responses are drained and their body included in the returned error.
- `http.NoBody` used for requests with no payload; `Content-Type` only set when
  a body is present.

## API Conventions

```go
// Required field
// +kubebuilder:validation:Required
Name string `json:"name"`

// Optional pointer field
// +optional
Description *string `json:"description,omitempty"`

// Enum-constrained field (User.Tags example)
// +kubebuilder:validation:Enum=administrator;monitoring;policymaker;management;impersonator
Tags []string `json:"tags,omitempty"`
```

User passwords are **never** stored in the spec. Use a `SecretKeySelector`:
```go
// +optional
PasswordSecretRef *xpv1.SecretKeySelector `json:"passwordSecretRef,omitempty"`
```

## Status Conditions

```go
xpv1.Available()  // resource exists and is up-to-date
xpv1.Creating()   // Create in progress
xpv1.Deleting()   // Delete in progress
```

## Error Handling

```go
errors.Wrap(err, "descriptive context")  // always wrap with context
clients.IsNotFound(err)                  // true when status 404
```
