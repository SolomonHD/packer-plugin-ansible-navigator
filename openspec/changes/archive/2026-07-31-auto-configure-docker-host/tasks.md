# Tasks: Auto-configure Docker Host Mapping

## 1. Implementation
- [x] Update `applyAutomaticEEDefaults` signature in `navigator_config.go` to accept `ansibleProxyHost`.
- [x] Implement detection logic: if `ansibleProxyHost == "gateway.docker.internal"`, ensure `--add-host=gateway.docker.internal:host-gateway` is in `ContainerOptions`.
- [x] Update `GenerateNavigatorConfigYAML` signature in `navigator_config.go`.
- [x] Update `provisioner.go` to pass `p.config.AnsibleProxyHost` to `GenerateNavigatorConfigYAML`.

## 2. Testing
- [x] Add unit test in `navigator_config_test.go` verifying the flag is added when proxy host matches.
- [x] Add unit test verifying the flag is NOT added when proxy host differs.
- [x] Add unit test verifying no duplicate flags are added if user already specified it.
