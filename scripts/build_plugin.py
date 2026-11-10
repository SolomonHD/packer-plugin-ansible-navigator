#!/usr/bin/env python3
"""Build a Packer-compatible development artifact without publishing a release."""

import argparse
import datetime
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile
import zipfile


ROOT = Path(__file__).resolve().parent.parent
NAME = "packer-plugin-ansible-navigator"


def output(argv, env=None):
    return subprocess.check_output(argv, cwd=ROOT, env=env, text=True).strip()


def build(args):
    version = (ROOT / "version/VERSION").read_text().strip()
    if not re.fullmatch(r"(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)", version):
        raise ValueError("version/VERSION must contain a canonical base SemVer")
    target = f"{args.os}/{args.arch}"
    if target not in output(["go", "tool", "dist", "list"]).splitlines():
        raise ValueError(f"unsupported Go target: {target}")
    host = json.loads(output(["go", "env", "-json", "GOHOSTOS", "GOHOSTARCH"]))
    native_env = dict(os.environ, GOOS=host["GOHOSTOS"], GOARCH=host["GOHOSTARCH"], CGO_ENABLED="0")
    # CI supplies the API from its verified native binary. Local builds can
    # discover it natively even if the caller exported a cross-compilation GOOS.
    api = args.api_version
    if api is None:
        api = json.loads(output(["go", "run", ".", "describe"], native_env))["api_version"]
    if not re.fullmatch(r"x[0-9]+\.[0-9]+", api):
        raise ValueError("invalid plugin API version")
    module = output(["go", "list", "-m"])
    revision = output(["git", "rev-parse", "HEAD"])
    tree = output(["git", "rev-parse", "HEAD^{tree}"])
    dirty = bool(output(["git", "status", "--porcelain", "--untracked-files=normal"]))
    timestamp = int(output(["git", "show", "-s", "--format=%ct", "HEAD"]))
    # Packer only accepts the canonical -dev prerelease, not arbitrary SHA
    # suffixes or build metadata. Keep source identity in build-info.json.
    version += "-dev"
    stem = f"{NAME}_v{version}_{api}_{args.os}_{args.arch}"
    binary_name = stem + (".exe" if args.os == "windows" else "")
    env = dict(os.environ, GOOS=args.os, GOARCH=args.arch, CGO_ENABLED="0")
    with tempfile.TemporaryDirectory(prefix="packer-plugin-build-") as temporary:
        binary = Path(temporary) / binary_name
        subprocess.run(
            ["go", "build", "-trimpath", "-buildvcs=false", "-ldflags",
             f"-s -w -X {module}/version.VersionPrerelease=dev", "-o", str(binary), "."],
            cwd=ROOT, env=env, check=True,
        )
        if target == f"{host['GOHOSTOS']}/{host['GOHOSTARCH']}":
            description = json.loads(output([str(binary), "describe"]))
            if description["version"] != version or description["api_version"] != api:
                raise ValueError("binary description does not match the package name")
            if set(description["provisioners"]) != {"-packer-default-plugin-name-", "local"}:
                raise ValueError("binary is missing a provisioner")
        destination = args.output_dir.resolve()
        destination.mkdir(parents=True, exist_ok=True)
        archive = destination / f"{stem}.zip"
        # Wrapping the executable in our own ZIP retains Unix executable mode
        # when GitHub's outer artifact archive resets file permissions.
        date = datetime.datetime.fromtimestamp(max(timestamp, 315532800), datetime.timezone.utc)
        with zipfile.ZipFile(archive, "w") as package:
            for name, data, mode in (
                (binary_name, binary.read_bytes(), 0o100755),
                ("LICENSE.txt", (ROOT / "LICENSE").read_bytes(), 0o100644),
            ):
                entry = zipfile.ZipInfo(name, date.timetuple()[:6])
                entry.create_system = 3
                entry.external_attr = mode << 16
                entry.compress_type = zipfile.ZIP_DEFLATED
                package.writestr(entry, data)
        archive_hash = hashlib.sha256(archive.read_bytes()).hexdigest()
        (destination / "SHA256SUMS").write_text(f"{archive_hash}  {archive.name}\n")
        metadata = {
            "plugin": NAME, "version": version, "api_version": api,
            "os": args.os, "arch": args.arch, "source_commit": revision,
            "source_tree": tree, "source_dirty": dirty,
            "go_version": output(["go", "version"]), "cgo_enabled": False,
            "binary": binary_name, "binary_sha256": hashlib.sha256(binary.read_bytes()).hexdigest(),
            "archive": archive.name, "archive_sha256": archive_hash,
        }
        if os.environ.get("GITHUB_RUN_ID"):
            metadata["workflow_run"] = os.environ["GITHUB_RUN_ID"]
        (destination / "build-info.json").write_text(json.dumps(metadata, indent=2) + "\n")
        print(archive)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--os", required=True, help="Go target OS, for example linux")
    parser.add_argument("--arch", required=True, help="Go target architecture, for example amd64")
    parser.add_argument("--api-version", help="API reported by a verified native build, e.g. x5.0")
    parser.add_argument("--output-dir", type=Path, default=ROOT / "dist")
    args = parser.parse_args()
    try:
        build(args)
    except (ValueError, subprocess.CalledProcessError) as error:
        parser.exit(1, f"Plugin build failed: {error}\n")


if __name__ == "__main__":
    main()
