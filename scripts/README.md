# Scripts

Automation scripts for provider-rabbitmq development and release management.

## Release Preparation

1. Update `VERSION`, `package/crossplane.yaml`, and current installation references.
2. Add the release entry to `CHANGELOG.md`.
3. Open a release PR from `release/v0.5.3` based on `origin/master`.
4. After merge, create and push the exact release tag:

   ```bash
   git tag v0.5.3
   git push origin v0.5.3
   ```

5. The tag-only workflow builds and publishes `linux_amd64` and `linux_arm64` packages, aliases `latest`, verifies equal digests and both architectures, and creates the GitHub Release.
