# Execution environment configuration verification

Verified November 9, 2026 on WSL2, branch `fix/execution-environment-config`, based on `5f1cd683d54055690ca8b9f66c23f8fb794526d5`.

## Delivered behavior

Both provisioners retain the existing public HCL fields and generated schema. Actual HCL parsing, generated-schema decoding, and SDK conversion tests cover engine selection, ordered container options and pull arguments, repeatable mount blocks, and rejection of mount-list assignment. YAML tests cover pull policy only, arguments only, both, and neither, plus custom mounts and automatic defaults.

The on-target runtime now recognizes pull arguments as unmapped settings and includes their ordered list in its existing minimal YAML upload. Communicator integration tests inspect that upload and its remote command together, including existing mapped CLI flags and settings path. They verify local temporary-file cleanup on success and run failure. Overlaying the original `navigator_config.go` makes the arguments-only, combined, and upload regressions fail; the updated source passes.

README, configuration reference, and the both-provisioner example describe the actual block syntax and host/target runtime prerequisites. The companion playbook uses `ansible.builtin.ping`.

## Validation

- `go test ./...`: pass (remote and local packages).
- `go vet ./...`: pass.
- `go build ./...`: pass.
- `make plugin-check`: pass, including source binary build and x5 check.
- `make generate`: pass; generated tracked files have no diff.
- `packer fmt -check example/execution-environment-config.pkr.hcl`: pass.
- `packer validate -var 'local_target_image=example/navigator-docker-target:validation' example/execution-environment-config.pkr.hcl`: pass for both named builds. No `packer build`, `packer init`, runtime image pull, or infrastructure change was performed.
- `openspec validate expand-execution-environment-config --strict`: pass.
- `openspec validate --specs`: all 11 main specs pass with warnings.
- `openspec validate --all --strict`: both active changes pass; the 11 unchanged main specs fail on 33 pre-existing warnings (placeholder Purpose sections, old scenario formatting, and long requirement descriptions). No validation errors were reported. At implementation verification time, main specs and the other active change were unmodified.
- `git diff --check`: pass.

Go commands used the existing temporary GOPATH/module/build caches; generation and plugin check also used `/tmp/navigator-go-bin`. The initial generation attempt encountered read-only default GOPATH sumdb storage; rerunning with the isolated GOPATH passed.

## Source plugin selection

Packer 1.14.2 validated using an isolated `PACKER_PLUGIN_PATH` with the prepared Docker builder 1.1.2 and this source-built plugin 8.7.0. The installed plugin's SHA256 matched the built binary:

`aec05567e434a394e9ef3904e55faabbdee834cdcc9c4eec554e479bacf6ad0b`

No plugin version bump or release is included. Go was 1.26.3. The installed Navigator version was 26.11.0; its local JSON schema confirms the nested pull argument list and mount-block output structure used in the regressions.

## Acceptance boundaries and follow-up

The prior WSL/native Linux records under `docs/acceptance/` describe the SSH provisioner's separately recorded source snapshots. They do not establish on-target live EE support, and no new live runtime qualification was run here because SSH runtime behavior is unchanged.

On-target tests exercise the plugin's configuration handoff using a communicator fixture. They preserve the existing engine flag and JSON container-option construction; those CLI mappings still require compatibility review against the selected Navigator release and an actual target runtime. Passing these tests or `packer validate` does not certify those existing mappings or prove Docker-on-target execution.

Review this implementation and its evidence before spec sync/archive or merge. At the end of apply, no commit, push, merge, archive, tag, or release had been performed for this change.

## Archive closeout

On November 9, 2026, under the subsequent archive-and-merge request, synced four advanced EE requirements into each main provisioner capability and extended the local configuration-generation requirement. Verified every delta block in its main spec, preservation of unrelated requirements and Purpose text, and successful non-strict main-spec validation (11/11). All artifacts and 7/7 tasks were complete before moving this change to `openspec/changes/archive/2026-11-09-expand-execution-environment-config/`. The independent `add-show-extra-vars-option` change remains active. GitHub squash/merge follows this archive; no release or tag is requested.
