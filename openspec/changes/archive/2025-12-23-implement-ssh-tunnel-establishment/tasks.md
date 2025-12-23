# Tasks: Implement SSH Tunnel Establishment

## Implementation Tasks

- [x] Add `setupSSHTunnel()` function to `provisioner.go`
  - [x] Function signature: `(ui packersdk.Ui, targetHost string, targetPort int) (localPort int, tunnel io.Closer, err error)`
  - [x] Parse bastion private key file using `ssh.ParsePrivateKey()` if specified
  - [x] Create password auth method using `ssh.Password()` if specified
  - [x] Establish SSH client connection to bastion host
  - [x] Implement dynamic port allocation (try up to 10 ports if `local_port` specified, else system-assigned)
  - [x] Create local TCP listener on `127.0.0.1:<allocatedPort>`
  - [x] Set up port forwarding goroutine to forward connections through bastion to target
  - [x] Return local port, closer handle, and error
  - [x] Add appropriate error wrapping for each failure mode

- [x] Integrate tunnel into `Provision()` workflow
  - [x] Add condition check for `p.config.SSHTunnelMode`
  - [x] Call `setupSSHTunnel()` when tunnel mode is true
  - [x] Store returned local port in `p.config.LocalPort`
  - [x] Override `generatedData["Host"]` to `"127.0.0.1"`
  - [x] Override `generatedData["Port"]` to tunnel local port
  - [x] Defer tunnel cleanup (`tunnel.Close()`) immediately after successful setup
  - [x] Skip `setupAdapter()` call when `ssh_tunnel_mode = true`
  - [x] Add UI messages for tunnel setup and cleanup
  - [x] Extract target host and port from `generatedData`
  - [x] Handle target credentials (use communicator SSH key for Ansible)

- [x] Add error handling and logging
  - [x] Bastion connection failure: "Failed to connect to bastion host <host>:<port>: <error>"
  - [x] Bastion auth failure: "Failed to authenticate to bastion: <error>"
  - [x] Invalid key format: "Failed to parse bastion private key: <error>"
  - [x] Target unreachable: "Failed to establish tunnel to target <host>:<port>: <error>"
  - [x] Port allocation failure: "Failed to allocate local port for tunnel"
  - [x] Add UI.Say() for tunnel setup start, success, and cleanup

## Validation Tasks

- [x] Run `go build ./...` to verify compilation
- [x] Run `go test ./...` to ensure no regressions
- [x] Run `make plugin-check` to validate plugin conformance
- [x] Verify error messages are clear and actionable

## Documentation Tasks

(These are tracked in change `04-update-documentation`, not this change)

- Configuration documentation update (out of scope)
- Examples for SSH tunnel usage (out of scope)
