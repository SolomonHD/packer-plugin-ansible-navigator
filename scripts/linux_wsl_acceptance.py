#!/usr/bin/env python3
"""Explicit, disposable real-Navigator acceptance. No live work at import time.

Raw child output is private and ephemeral. Only allowlisted evidence is exported.
Preparation downloads inputs; run never installs remote plugins or pulls images.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path
import platform
import re
import shutil
import shlex
import signal
import subprocess
import sys
import tempfile
import time
import uuid
from dataclasses import dataclass
from urllib.parse import urlsplit

ROOT = Path(__file__).resolve().parents[1]
FIXTURE = ROOT / "provisioner/ansible-navigator/test-fixtures/linux-wsl"
LABEL = "io.packer.navigator.acceptance"
NAV_SOURCE = "github.com/solomonhd/ansible-navigator"
BUILDER_SOURCE = "github.com/hashicorp/docker"


@dataclass
class Result:
    code: int
    output: str = ""
    timed_out: bool = False


class Failed(Exception):
    def __init__(self, stage, reason, result=None):
        super().__init__(f"{stage}: {reason}")
        self.stage, self.reason, self.result = stage, reason, result


def execute(argv, *, env=None, cwd=None, timeout=30):
    """Kill the complete child group on timeout/interruption; never stream output."""
    with tempfile.TemporaryFile(mode="w+b") as output:
        proc = subprocess.Popen([str(a) for a in argv], env=env, cwd=cwd,
                                stdin=subprocess.DEVNULL, stdout=output,
                                stderr=subprocess.STDOUT, start_new_session=True)
        timed_out = False
        try:
            proc.wait(timeout=timeout)
        except subprocess.TimeoutExpired:
            timed_out = True
            os.killpg(proc.pid, signal.SIGKILL)
            proc.wait(timeout=5)
        except BaseException:
            try:
                os.killpg(proc.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
            proc.wait(timeout=5)
            raise
        output.seek(0)
        # Keep a finite private diagnostic window; never publish raw output.
        data = output.read(8 * 1024 * 1024).decode("utf-8", errors="replace")
    return Result(proc.returncode, data, timed_out)


def digest(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def require(result, stage, reason):
    if result.timed_out:
        raise Failed(stage, "timeout", result)
    if result.code:
        raise Failed(stage, reason, result)
    return result.output


def endpoint_description(env, context_result=None):
    """Keep endpoint kind and selection source, never hosts, paths or userinfo."""
    source = "DOCKER_CONTEXT" if env.get("DOCKER_CONTEXT") else (
        "DOCKER_HOST" if env.get("DOCKER_HOST") else "active-context")
    endpoint = env.get("DOCKER_HOST", "")
    if context_result and not context_result.code:
        endpoint = context_result.output.strip()
    kind = urlsplit(endpoint).scheme if endpoint else "unknown"
    return {"source": source, "endpoint_kind": kind if kind in (
        "unix", "npipe", "tcp", "ssh", "http", "https") else "unknown",
        "docker_config_explicit": bool(env.get("DOCKER_CONFIG")),
        "tls_verify_explicit": bool(env.get("DOCKER_TLS_VERIFY")),
        "cert_path_explicit": bool(env.get("DOCKER_CERT_PATH"))}


def failure_reason(stage, result):
    if result.timed_out:
        return "timeout"
    text = result.output.lower()
    if stage == "daemon":
        if "permission denied" in text or "access is denied" in text:
            return "endpoint-permission-denied"
        if "no such file" in text or "cannot connect" in text or "connection refused" in text:
            return "endpoint-unavailable"
        return "daemon-probe-failed"
    if "no such image" in text or "image not known" in text or "not present" in text:
        return "image-unavailable"
    return "ee-startup-failed"


def valid_start(events, run_id, image):
    for line in events.splitlines():
        try:
            event = json.loads(line)
        except (ValueError, TypeError):
            continue
        if not isinstance(event, dict) or not isinstance(event.get("Actor"), dict):
            continue
        attrs = event["Actor"].get("Attributes", {})
        if not isinstance(attrs, dict):
            continue
        if (event.get("Type") == "container" and
                event.get("Action", event.get("status")) == "start" and
                attrs.get(LABEL) == run_id and attrs.get("image") == image and
                re.fullmatch(r"[0-9a-f]{12,64}", event.get("Actor", {}).get("ID", ""))):
            return event["Actor"]["ID"]
    return None


def token_seen(output, token):
    # Packer prefixes streamed shell output with the builder name. Do not count
    # an echoed command, config dump, recap or token embedded in another value.
    clean = re.sub(r"\x1b\[[0-9;]*m", "", output)
    pattern = re.compile(r"^(?:==>\s*)?(?:[A-Za-z0-9_.-]+:\s*)?" + re.escape(token) + r"$")
    return any(pattern.fullmatch(line.strip()) for line in clean.splitlines())


def verify_build(result, nonce, events, run_id, image, negative=False):
    if result.timed_out:
        raise Failed("negative-control" if negative else "provisioning", "timeout", result)
    absent = token_seen(result.output, f"ACCEPTANCE_ABSENT:{nonce}")
    verified = token_seen(result.output, f"ACCEPTANCE_VERIFIED:{nonce}")
    launched = valid_start(events, run_id, image)
    if negative:
        if (result.code == 0 or not absent or verified or not launched or
                f"ACCEPTANCE_INTENTIONAL_FAILURE:{nonce}" not in result.output):
            raise Failed("negative-control", "expected-task-failure-not-proved", result)
    elif result.code != 0 or not absent or not verified or not launched:
        if result.code:
            stage, reason = provisioning_failure(result.output)
        else:
            stage, reason = "target-verification", "independent-target-proof-missing"
        raise Failed(stage, reason, result)
    return {"packer_exit": result.code, "marker_absent": absent,
            "target_verified": verified, "ee_container_id": launched,
            "intentional_failure_observed": negative}


def provisioning_failure(output):
    """Classify private diagnostics without exporting file paths or task data."""
    text = output.lower()
    if any(pattern in text for pattern in ("unreachable!", "failed to connect to the host via ssh",
                                           "connection refused", "permission denied (publickey")):
        return "ssh-proxy", "target-ssh-connection-failed"
    if any(pattern in text for pattern in ("unable to retrieve file contents", "could not find or access",
                                           "could not be found", "no such file or directory")):
        return "ee-file-visibility", "required-execution-file-unavailable"
    if "failed!" in text or "acceptance_intentional_failure:" in text:
        return "ansible-task", "target-task-failed"
    return "provisioning", "packer-build-failed"


def load_profile(path):
    profile = json.loads(path.read_text())
    for key in ("target_image", "ee_image"):
        if not re.fullmatch(r"[A-Za-z0-9._/:\-]+@sha256:[0-9a-f]{64}", profile.get(key, "")):
            raise Failed("profile", f"{key}-requires-immutable-digest")
    if not re.fullmatch(r"[0-9]+\.[0-9]+\.[0-9]+", profile.get("docker_builder_version", "")):
        raise Failed("profile", "exact-docker-builder-version-required")
    if profile.get("architecture") not in ("linux/amd64", "linux/arm64"):
        raise Failed("profile", "supported-image-platform-required")
    return profile


def executable(command, env):
    selected = shutil.which(command, path=env.get("PATH", ""))
    if not selected:
        raise Failed("executable", "required-executable-missing")
    return str(Path(selected).absolute())


def real_navigator(selected, env, run=execute):
    """Get explicit baseline path without treating a manager shim as real binary."""
    path = Path(selected)
    if path.parent.name == "shims":
        for manager in ("asdf", "rbenv", "pyenv"):
            if manager in str(path.parent.parent):
                binary = executable(manager, env)
                value = require(run([binary, "which", path.name], env=env, timeout=10),
                                "executable", "shim-resolution-failed").strip()
                candidate = Path(value)
                if (not candidate.is_absolute() or candidate.parent.name == "shims" or
                        not candidate.is_file() or not os.access(candidate, os.X_OK)):
                    raise Failed("executable", "invalid-shim-result")
                return str(candidate)
    return str(path.resolve())


class Runner:
    def __init__(self, args, run=execute):
        self.args, self.run = args, run
        self.env = dict(os.environ)
        # Parent settings select the endpoint. No sudo/context switch/socket fix.
        self.env.update({"NO_COLOR": "1", "ANSIBLE_FORCE_COLOR": "0", "PACKER_NO_COLOR": "1"})
        self.profile = load_profile(args.profile)
        self.run_id = "navacc-" + uuid.uuid4().hex
        self.owned = []
        self.temp = None
        self.work = None
        self.docker = None
        self.record = {"format_version": 1, "run_id": self.run_id,
                       "evidence_kind": "live", "result": "failed", "stages": [],
                       "profile": self.profile, "profile_sha256": digest(args.profile),
                       "host": {"hostname": platform.node(), "kernel": platform.release(),
                                "system": platform.system(), "machine": platform.machine()},
                       "variants": []}

    def call(self, argv, timeout=30, env=None, cwd=None):
        return self.run(argv, env=self.env if env is None else env,
                        cwd=cwd or self.work or self.temp or ROOT, timeout=timeout)

    def stage(self, name, **evidence):
        self.record["stages"].append({"name": name, "status": "passed", **evidence})
        print(f"{name}: passed", flush=True)

    def binaries(self):
        self.docker = executable("docker", self.env)
        self.packer = real_navigator(executable("packer", self.env), self.env, self.run)
        self.go = executable("go", self.env)
        selected = self.args.navigator or "ansible-navigator"
        self.navigator_selected = executable(selected, self.env)
        self.navigator = real_navigator(self.navigator_selected, self.env, self.run)
        version = require(self.call([self.navigator, "--version"], 10),
                          "executable", "navigator-version-failed")
        manager = next((name for name in ("asdf", "rbenv", "pyenv")
                        if Path(self.navigator_selected).parent.name == "shims" and
                        name in str(Path(self.navigator_selected).parent.parent)), None)
        self.stage("executable", navigator_path=self.navigator, shim_manager=manager,
                   inherited_selection=self.navigator_selected,
                   navigator_version=safe_version(version))

    def daemon(self, env=None, docker=None, variant=None):
        env = self.env if env is None else env
        docker = self.docker if docker is None else docker
        context = self.call([docker, "context", "inspect", "--format",
                             "{{.Endpoints.docker.Host}}"], 10, env)
        selection = endpoint_description(env, context)
        self.record["docker_selection"] = selection
        response = self.call([docker, "version", "--format", "{{json .Server.Version}}"], 30, env)
        if response.code or response.timed_out:
            raise Failed("daemon", failure_reason("daemon", response), response)
        version = response.output.strip().strip('"')
        if not re.fullmatch(r"[0-9][A-Za-z0-9.+\-]{0,80}", version):
            raise Failed("daemon", "unexpected-daemon-version-response")
        # Qualification is local Unix only; do not silently qualify a remote daemon.
        if self.record["docker_selection"]["endpoint_kind"] != "unix":
            raise Failed("daemon", "local-unix-endpoint-required")
        client = require(self.call([docker, "version", "--format", "{{.Client.Version}}"], 10, env),
                         "daemon", "docker-client-version-failed")
        self.stage("daemon", docker_server_version=version, docker_client_version=safe_version(client),
                   variant=variant, selection=selection, docker_executable=docker)

    def images(self):
        images = {}
        for key in ("target_image", "ee_image"):
            image = self.profile[key]
            result = self.call([self.docker, "image", "inspect", "--format",
                                "{{json .Id}} {{json .Architecture}} {{json .Os}}", image], 10)
            if result.code or result.timed_out:
                raise Failed("images", "prepared-image-unavailable" if not result.timed_out else "timeout", result)
            parts = result.output.strip().split()
            if len(parts) != 3:
                raise Failed("images", "unexpected-image-metadata")
            identity, arch, system = [json.loads(x) for x in parts]
            if f"{system}/{arch}" != self.profile["architecture"]:
                raise Failed("images", "prepared-image-platform-mismatch")
            if not re.fullmatch(r"sha256:[0-9a-f]{64}", identity):
                raise Failed("images", "invalid-image-identity")
            images[key] = {"id": identity, "architecture": arch, "os": system}
        self.stage("images", images=images)

    def event_log(self, run_id, since, env=None, docker=None):
        # Fixed until makes docker events exit. The subprocess remains bounded.
        until = f"{time.time():.9f}"
        result = self.call([docker or self.docker, "events", "--since", str(since), "--until", until,
                            "--filter", "type=container", "--filter", "event=start",
                            "--filter", f"label={LABEL}={run_id}", "--format", "{{json .}}"], 10, env)
        return require(result, "ee-evidence", "event-collection-failed")

    def startup(self, env=None, navigator=None, docker=None, variant=None):
        env = dict(self.env if env is None else env)
        navigator = navigator or self.navigator
        docker = docker or self.docker
        startup_id = f"{self.run_id}-{variant}-startup" if variant else self.run_id
        nonce = uuid.uuid4().hex
        settings = {"ansible-navigator": {"mode": "stdout", "execution-environment": {
            "enabled": True, "container-engine": "docker", "image": self.profile["ee_image"],
            "pull": {"policy": "never"}, "container-options": [f"--label={LABEL}={startup_id}", "--tty=false"],
            "environment-variables": {"set": {"HOME": "/tmp", "ANSIBLE_LOCAL_TMP": "/tmp/.ansible-local"}}},
            "logging": {"file": str(self.temp / "navigator-startup.log"), "level": "warning"}}}
        config = self.temp / "startup.json"
        config.write_text(json.dumps(settings))  # JSON is a YAML subset.
        env["ANSIBLE_NAVIGATOR_CONFIG"] = str(config)
        payload = ("import importlib.metadata as m,json,platform,subprocess;"
                   f"print('ACCEPTANCE_STARTUP:{nonce}');"
                   "print('ACCEPTANCE_VERSIONS:'+json.dumps({'python':platform.python_version(),"
                   "'ansible_core':m.version('ansible-core'),'runner':m.version('ansible-runner'),"
                   "'ssh':subprocess.check_output(['ssh','-V'],stderr=subprocess.STDOUT,text=True).strip()}))")
        since = f"{time.time():.9f}"
        self.owned.append(startup_id)
        # Navigator exec defaults to shell=True and joins its pass-through
        # arguments. Quote the payload once for that shell boundary.
        result = self.call([navigator, "exec", "--", shlex.join(["python3", "-c", payload])], 120, env)
        if result.code or result.timed_out:
            raise Failed("ee-startup", failure_reason("ee-startup", result), result)
        events = self.event_log(startup_id, since, env, docker)
        container = valid_start(events, startup_id, self.profile["ee_image"])
        if not token_seen(result.output, f"ACCEPTANCE_STARTUP:{nonce}") or not container:
            raise Failed("ee-startup", "correlated-launch-proof-missing", result)
        versions = None
        for line in result.output.splitlines():
            if line.startswith("ACCEPTANCE_VERSIONS:"):
                try:
                    versions = json.loads(line.split(":", 1)[1])
                except ValueError:
                    pass
        if not isinstance(versions, dict) or not versions.get("ssh"):
            raise Failed("ee-startup", "required-ee-tools-not-proved")
        allowed = {key: safe_version(str(versions.get(key, ""))) for key in ("python", "ansible_core", "runner", "ssh")}
        if any(value == "unrecognized" for value in allowed.values()):
            raise Failed("ee-startup", "required-ee-versions-not-proved")
        self.stage("ee-startup", ee_container_id=container, startup_nonce=nonce, versions=allowed,
                   variant=variant, navigator_executable=navigator, docker_executable=docker)

    def install_source(self):
        plugin_dir = self.temp / "plugins"
        plugin_dir.mkdir(mode=0o700)
        prepared = self.args.prepared_dir / "plugins" / BUILDER_SOURCE
        if not prepared.is_dir() or prepared.is_symlink():
            raise Failed("source-plugin", "prepared-builder-missing")
        # The only prepared subtree admitted is the exact Docker builder.
        builder_files = list(prepared.glob(f"packer-plugin-docker_v{self.profile['docker_builder_version']}_*"))
        binaries = [p for p in builder_files if p.is_file() and not p.name.endswith("_SHA256SUM")]
        if len(binaries) != 1:
            raise Failed("source-plugin", "exact-prepared-builder-required")
        destination = plugin_dir / BUILDER_SOURCE
        destination.mkdir(parents=True)
        for path in builder_files:
            if path.is_symlink() or not path.is_file():
                raise Failed("source-plugin", "unsafe-prepared-builder-file")
            shutil.copy2(path, destination / path.name)
        self.env["PACKER_PLUGIN_PATH"] = str(plugin_dir)
        self.env["PACKER_CACHE_DIR"] = str(self.temp / "packer-cache")
        binary = self.temp / "packer-plugin-ansible-navigator"
        require(self.call([self.go, "build", "-o", binary, "."], 180, cwd=ROOT),
                "source-plugin", "source-build-failed")
        require(self.call([self.packer, "plugins", "install", "--path", binary, NAV_SOURCE], 30),
                "source-plugin", "local-plugin-install-failed")
        installed = list((plugin_dir / NAV_SOURCE).glob("packer-plugin-ansible-navigator_*"))
        installed = [p for p in installed if p.is_file() and not p.name.endswith("_SHA256SUM")]
        if len(installed) != 1 or digest(installed[0]) != digest(binary):
            raise Failed("source-plugin", "installed-source-checksum-mismatch")
        self.installed_plugin = installed[0]
        self.source_plugin_checksum = digest(binary)
        self.stage("source-plugin", binary_sha256=self.source_plugin_checksum, installed_path=str(installed[0]),
                   docker_builder_version=self.profile["docker_builder_version"])

    def variant_env(self, variant):
        env = dict(self.env)
        if variant == "explicit":
            return env, self.navigator, []
        if variant == "path":
            return env, Path(self.navigator_selected).name, []
        # Preserve required tools while preventing name discovery in inherited PATH.
        tools = self.temp / "restricted-tools"
        tools.mkdir(exist_ok=True)
        names = ("docker", "python3", "python", "env", "sh", "bash", "cat", "chmod", "mkdir", "rm", "ssh")
        for name in names:
            source = sys.executable if name in ("python", "python3") else shutil.which(name, path=self.env.get("PATH", ""))
            if source and not (tools / name).exists():
                (tools / name).symlink_to(source)
        env["PATH"] = str(tools)
        if shutil.which(Path(self.navigator).name, path=env["PATH"]):
            raise Failed("configured-path", "navigator-visible-in-restricted-parent-path")
        return env, Path(self.navigator).name, [str(Path(self.navigator).parent)]

    def probe_variant(self, variant):
        inherited, command, search = self.variant_env(variant)
        effective = dict(inherited)
        if search:
            effective["PATH"] = os.pathsep.join([*search, inherited.get("PATH", "")])
        navigator = real_navigator(executable(command, effective), effective, self.run)
        docker = executable("docker", effective)
        self.daemon(effective, docker, variant)
        self.startup(effective, navigator, docker, variant)

    def build(self, variant, negative=False):
        env, command, search = self.variant_env(variant)
        build_id = f"{self.run_id}-{variant}-{'negative' if negative else 'positive'}"
        self.owned.append(build_id)
        nonce = uuid.uuid4().hex
        metadata = {"name": variant, "build_id": build_id, "negative_control": negative,
            "nonce": nonce, "navigator_command": command, "navigator_path": search,
            "path_source": "restricted-tools+configured" if variant == "configured-path" else variant,
            "docker_selection": self.record.get("docker_selection", endpoint_description(env)),
            "version_check_enabled": True, "status": "started"}
        # Record identity before validation or launch, so unsuccessful variants
        # remain attributable rather than disappearing from the evidence.
        self.record["variants"].append(metadata)
        values = {"navigator_command": command, "navigator_path": search,
                  "target_image": self.profile["target_image"], "ee_image": self.profile["ee_image"],
                  "run_id": build_id, "marker_nonce": nonce, "expect_failure": negative}
        try:
            variables = self.temp / f"{build_id}.pkrvars.json"
            variables.write_text(json.dumps(values))
            template = FIXTURE / "acceptance.pkr.hcl"
            validate_args = [self.packer, "validate", "-var-file", variables, template]
            require(self.call(validate_args, 30, env), "validation", "packer-validate-failed")
            self.stage("validation", variant=variant, build_id=build_id, negative_control=negative)
            selected = require(self.call([self.packer, "plugins", "required", template], 10, env),
                               "source-plugin", "plugin-selection-query-failed")
            checksum = digest(self.installed_plugin)
            if str(self.installed_plugin) not in selected or checksum != self.source_plugin_checksum:
                raise Failed("source-plugin", "packer-selected-path-not-source-plugin")
            metadata.update(selected_plugin_path=str(self.installed_plugin), plugin_sha256=checksum)
            self.stage("source-plugin-selection", variant=variant, build_id=build_id,
                       selected_path=str(self.installed_plugin), binary_sha256=checksum)
            since = f"{time.time():.9f}"
            result = self.call([self.packer, "build", "-on-error=abort", "-var-file", variables, template], 600, env)
            events = self.event_log(build_id, since)
            proof = verify_build(result, nonce, events, build_id, self.profile["ee_image"], negative)
            metadata.update(status="passed", **proof)
            self.stage(f"{variant}-{'negative-control' if negative else 'provisioning'}", **proof)
        except Failed as error:
            metadata.update(status="failed", failed_stage=error.stage, reason=error.reason)
            if error.result:
                metadata.update(packer_exit=error.result.code, timed_out=error.result.timed_out)
            raise
        except KeyboardInterrupt:
            metadata.update(status="failed", failed_stage="interruption", reason="interrupted")
            raise
        except OSError:
            metadata.update(status="failed", failed_stage="runner", reason="local-operation-failed")
            raise

    def cleanup(self):
        errors = []
        if self.docker:
            for owned in self.owned:
                try:
                    ids = self.call([self.docker, "ps", "-aq", "--filter", f"label={LABEL}={owned}"], 10)
                    if ids.code or ids.timed_out:
                        errors.append("owned-container-query-failed")
                        continue
                    identities = ids.output.split()
                    if any(not re.fullmatch(r"[0-9a-f]{12,64}", item) for item in identities):
                        errors.append("invalid-owned-container-identity")
                        continue
                    if identities:
                        removed = self.call([self.docker, "rm", "-f", *identities], 30)
                        if removed.code or removed.timed_out:
                            errors.append("owned-container-removal-failed")
                    remaining = self.call([self.docker, "ps", "-aq", "--filter", f"label={LABEL}={owned}"], 10)
                    if remaining.code or remaining.timed_out or remaining.output.strip():
                        errors.append("owned-container-cleanup-unproved")
                except (OSError, Failed):
                    errors.append("owned-container-cleanup-error")
        if self.temp:
            try:
                shutil.rmtree(self.temp)
                if self.temp.exists():
                    errors.append("private-directory-remains")
            except OSError:
                errors.append("private-directory-removal-failed")
            self.temp = None
        self.record["cleanup"] = {"status": "failed" if errors else "passed", "errors": sorted(set(errors)),
                                  "shared_images_removed": False}
        return not errors

    def identity(self):
        sha = require(self.call(["git", "rev-parse", "HEAD"], 10, cwd=ROOT), "identity", "git-sha-failed").strip()
        dirty = require(self.call(["git", "status", "--porcelain"], 10, cwd=ROOT), "identity", "git-status-failed")
        files = sorted(set(list(ROOT.rglob("*.go")) + [ROOT / "go.mod", ROOT / "go.sum"] +
                           list((ROOT / "version").glob("*")) + list(FIXTURE.glob("*")) + [Path(__file__)]))
        combined = hashlib.sha256()
        for path in files:
            if path.is_file() and ".git" not in path.parts:
                combined.update(str(path.relative_to(ROOT)).encode() + b"\0" + path.read_bytes())
        self.record["source"] = {"sha": sha, "dirty": bool(dirty), "source_fixture_sha256": combined.hexdigest(),
            "fixture_sha256": {p.name: digest(p) for p in FIXTURE.glob("*") if p.is_file()}}
        self.record["reproduction"] = [sys.executable, str(Path(__file__).resolve()), "run", "--prepared-dir",
            str(self.args.prepared_dir), "--profile", str(self.args.profile), "--record", str(self.args.record),
            "--navigator", self.args.navigator or "ansible-navigator", *sum((["--variant", v] for v in self.args.variant), [])]
        self.record["host"]["python"] = platform.python_version()
        self.record["packer_version"] = safe_version(require(self.call([self.packer, "version"], 10),
                                                        "identity", "packer-version-failed"))

    def accept(self):
        self.temp = Path(tempfile.mkdtemp(prefix=self.run_id + "-"))
        os.chmod(self.temp, 0o700)
        self.work = self.temp / "work"
        self.work.mkdir(mode=0o700)
        temporary = self.temp / "private-tmp"
        temporary.mkdir(mode=0o700)
        # Keep generated credential files owned by the run even after a killed
        # build. TMPDIR is outside the working directory, so Navigator's normal
        # cwd mount cannot hide an EE file-visibility defect.
        self.env["TMPDIR"] = str(temporary)
        try:
            self.binaries()
            self.identity()
            self.daemon()
            self.images()
            self.install_source()
            for variant in self.args.variant:
                self.probe_variant(variant)
                self.build(variant)
                self.build(variant, negative=True)
            self.record["result"] = "passed"
        except Failed as error:
            item = {"name": error.stage, "status": "failed", "reason": error.reason}
            if error.result:
                item.update(exit_code=error.result.code, timed_out=error.result.timed_out)
            self.record["stages"].append(item)
            print(f"{error.stage}: failed ({error.reason})", file=sys.stderr)
        except KeyboardInterrupt:
            self.record["stages"].append({"name": "interruption", "status": "failed", "reason": "interrupted"})
        except OSError:
            self.record["stages"].append({"name": "runner", "status": "failed", "reason": "local-operation-failed"})
        finally:
            if not self.cleanup():
                self.record["result"] = "failed"
            self.args.record.parent.mkdir(parents=True, exist_ok=True)
            self.args.record.write_text(json.dumps(self.record, indent=2) + "\n")
        return 0 if self.record["result"] == "passed" else 1


def safe_version(text):
    # Versions have no reason to carry arbitrary child output into committed evidence.
    found = re.search(r"(?:\b|(?<=_))(?:v)?(\d+\.\d+(?:\.\d+)?(?:[A-Za-z0-9.+\-]{0,40}))", text)
    return found.group(1) if found else "unrecognized"


def prepare(args):
    profile = load_profile(args.profile)
    env = dict(os.environ)
    docker, packer = executable("docker", env), executable("packer", env)
    args.prepared_dir.mkdir(parents=True, exist_ok=True, mode=0o700)
    env["PACKER_PLUGIN_PATH"] = str(args.prepared_dir / "plugins")
    for key in ("target_image", "ee_image"):
        require(execute([docker, "pull", "--platform", profile["architecture"], profile[key]], env=env,
                        timeout=600), "preparation", f"{key}-pull-failed")
    require(execute([packer, "plugins", "install", BUILDER_SOURCE,
                     "=" + profile["docker_builder_version"]], env=env, timeout=180),
            "preparation", "docker-builder-install-failed")
    (args.prepared_dir / "profile.json").write_text(json.dumps(profile, indent=2) + "\n")
    print("preparation: passed (shared prepared images retained; isolated builder installed)")
    return 0


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("action", choices=("prepare", "run"))
    parser.add_argument("--profile", type=Path, default=FIXTURE / "profile.json")
    parser.add_argument("--prepared-dir", type=Path, required=True)
    parser.add_argument("--record", type=Path, default=Path("acceptance-record.json"))
    parser.add_argument("--navigator", help="Navigator name or explicit executable path")
    parser.add_argument("--variant", action="append", choices=("explicit", "configured-path", "path"))
    args = parser.parse_args(argv)
    args.profile, args.prepared_dir, args.record = [p.resolve() for p in (args.profile, args.prepared_dir, args.record)]
    args.variant = args.variant or ["explicit", "configured-path", "path"]
    if len(set(args.variant)) != len(args.variant):
        parser.error("duplicate variants are not allowed")
    try:
        return prepare(args) if args.action == "prepare" else Runner(args).accept()
    except Failed as error:
        print(str(error), file=sys.stderr)
        return 1


if __name__ == "__main__":
    signal.signal(signal.SIGTERM, lambda signum, frame: (_ for _ in ()).throw(KeyboardInterrupt()))
    sys.exit(main())
