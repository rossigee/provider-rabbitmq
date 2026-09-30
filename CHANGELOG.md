# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.5.3] - 2026-09-30

### Fixed

- **Decode RabbitMQ `tags` as a JSON array rather than a comma-separated string.** `GetUser`
  decoded the response into a `Tags string` field, but RabbitMQ returns an array, so every
  `User` reconcile failed with `cannot unmarshal array into Go struct field .tags of type
  string`. `splitTags` now takes `json.RawMessage` and accepts the array form, the CSV form,
  an absent field and `null`, returning nil instead of erroring on an unrecognised shape.
  `GetVhost` carried the identical string-typed field and is fixed in the same change; it had
  not yet been reached because the released image did not unmarshal the vhost response at all.
  This is a read-path fix — `CreateUser` and `CreateVhost` continue to send tags as a joined
  string, which RabbitMQ accepts on `PUT`.
- **`providerconfig` controller no longer drops the namespace** when re-reading a ProviderConfig
  for its status update. It used `client.ObjectKey{Name: pc.GetName()}`, which resolves in the
  empty namespace, so a namespaced ProviderConfig was reported as not found despite existing.
  Cosmetic in effect — it affected only the `Available` condition, not reconciliation.

### Changed

- Refreshed the Crossplane APIs fork dependency to `v2.5.0-rc.0`.
- Standardised the release workflow on the template applied in `f0149f4`, publishing the
  runtime image and `linux_amd64`/`linux_arm64` xpkg artifacts, aliasing the version as
  `latest`, and verifying digests and both platform manifests before creating the GitHub Release.
