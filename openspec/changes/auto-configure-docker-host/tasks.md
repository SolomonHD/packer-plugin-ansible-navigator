# Tasks: Auto-configure Docker Host Mapping

## 1. Implementation
- [ ] Update `applyAutomaticEEDefaults` signature in `navigator_config.go` to accept `ansibleProxyHost`.
- [ ] Implement detection logic: if `ansibleProxyHost == "gateway.docker.internal"`, ensure `--add-host=gateway.docker.internal:host-gateway` is in `ContainerOptions`.
- [ ] Update `GenerateNavigatorConfigYAML` signature in `navigator_config.go`.
- [ ] Update `provisioner.go` to pass `p.config.AnsibleProxyHost` to `GenerateNavigatorConfigYAML`.

## 2. Testing
- [ ] Add unit test in `navigator_config_test.go` verifying the flag is added when proxy host matches.
- [ ] Add unit test verifying the flag is NOT added when proxy host differs.
- [ ] Add unit test verifying no duplicate flags are added if user already specified it.
