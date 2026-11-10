"""Offline regression fixtures; none of these tests establish live acceptance."""
import argparse
import importlib.util
import json
import os
from pathlib import Path
import signal
import sys
import tempfile
import time
import unittest
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parent))
import linux_wsl_acceptance as acceptance

IMAGE = "public.invalid/ee@sha256:" + "a" * 64
TARGET = "public.invalid/target@sha256:" + "b" * 64
RUN = "navacc-test"
NONCE = "nonce-123"
CONTAINER = "1" * 64


def event(run=RUN, image=IMAGE, action="start"):
    return json.dumps({"Type": "container", "Action": action,
                       "Actor": {"ID": CONTAINER, "Attributes": {
                           acceptance.LABEL: run, "image": image}}})


class ProofTests(unittest.TestCase):
    def test_real_launch_requires_label_image_and_start_event(self):
        self.assertEqual(acceptance.valid_start(event(), RUN, IMAGE), CONTAINER)
        for invalid in (event(run="unrelated"), event(image=TARGET), event(action="create"),
                        '{"Server":{"Version":"29.0.0"}}', "not json", "null", "[]", '{"Actor":null}', '{"Actor":{"Attributes":null}}'):
            self.assertIsNone(acceptance.valid_start(invalid, RUN, IMAGE))

    def test_recap_host_marker_stale_marker_and_wrong_value_cannot_pass(self):
        for output in ("PLAY RECAP failed=0", f"ACCEPTANCE_VERIFIED:stale\nACCEPTANCE_ABSENT:{NONCE}",
                       f"ACCEPTANCE_VERIFIED:{NONCE}", f"ACCEPTANCE_ABSENT:{NONCE}",
                       f"echo ACCEPTANCE_ABSENT:{NONCE}\necho ACCEPTANCE_VERIFIED:{NONCE}"):
            with self.assertRaises(acceptance.Failed):
                acceptance.verify_build(acceptance.Result(0, output), NONCE, event(), RUN, IMAGE)

    def test_success_requires_exit_marker_and_correlated_ee(self):
        output = f"ACCEPTANCE_ABSENT:{NONCE}\nACCEPTANCE_VERIFIED:{NONCE}"
        proof = acceptance.verify_build(acceptance.Result(0, output), NONCE, event(), RUN, IMAGE)
        self.assertTrue(proof["target_verified"])
        for result, events in ((acceptance.Result(1, output), event()),
                               (acceptance.Result(0, output), event(run="another")),
                               (acceptance.Result(-9, output, True), event())):
            with self.assertRaises(acceptance.Failed):
                acceptance.verify_build(result, NONCE, events, RUN, IMAGE)

    def test_negative_control_requires_intentional_failure_and_no_success(self):
        output = f"ACCEPTANCE_ABSENT:{NONCE}\nACCEPTANCE_INTENTIONAL_FAILURE:{NONCE}"
        proof = acceptance.verify_build(acceptance.Result(1, output), NONCE, event(), RUN, IMAGE, True)
        self.assertFalse(proof["target_verified"])
        for result in (acceptance.Result(0, output), acceptance.Result(1, "connection refused"),
                       acceptance.Result(1, output + f"\nACCEPTANCE_VERIFIED:{NONCE}")):
            with self.assertRaises(acceptance.Failed):
                acceptance.verify_build(result, NONCE, event(), RUN, IMAGE, True)

    def test_failure_classification_separates_boundaries(self):
        cases = [("daemon", "permission denied", "endpoint-permission-denied"),
                 ("daemon", "Cannot connect to daemon", "endpoint-unavailable"),
                 ("ee-startup", "No such image", "image-unavailable"),
                 ("ee-startup", "unknown failure", "ee-startup-failed")]
        for stage, output, reason in cases:
            self.assertEqual(acceptance.failure_reason(stage, acceptance.Result(1, output)), reason)
        self.assertEqual(acceptance.failure_reason("daemon", acceptance.Result(-9, "", True)), "timeout")

    def test_provisioning_failure_preserves_stage_without_private_data(self):
        for output, stage in (("UNREACHABLE! CANARY", "ssh-proxy"),
                              ("Could not find or access /private/key CANARY", "ee-file-visibility"),
                              ("fatal: [default]: FAILED! secret=CANARY", "ansible-task"),
                              ("unknown CANARY", "provisioning")):
            result = acceptance.Result(1, output)
            with self.assertRaises(acceptance.Failed) as found:
                acceptance.verify_build(result, NONCE, "", RUN, IMAGE)
            self.assertEqual(found.exception.stage, stage)
            self.assertNotIn("CANARY", str(found.exception))

    def test_endpoint_evidence_does_not_retain_secrets(self):
        env = {"DOCKER_HOST": "tcp://username:CANARY@private.example:2376",
               "DOCKER_CONTEXT": "SECRET_CONTEXT", "DOCKER_CONFIG": "/private/registry/config",
               "DOCKER_TLS_VERIFY": "1", "DOCKER_CERT_PATH": "/private/key"}
        record = acceptance.endpoint_description(env)
        self.assertEqual(record["source"], "DOCKER_CONTEXT")
        encoded = json.dumps(record)
        for secret in ("CANARY", "username", "private", "SECRET_CONTEXT"):
            self.assertNotIn(secret, encoded)

    def test_version_output_is_allowlisted(self):
        self.assertEqual(acceptance.safe_version("ansible-navigator 26.11.0\nCANARY=secret"), "26.11.0")
        self.assertEqual(acceptance.safe_version("SECRET=canary"), "unrecognized")


class RunnerTests(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.base = Path(self.directory.name)
        self.profile = self.base / "profile.json"
        self.profile.write_text(json.dumps({"target_image": TARGET, "ee_image": IMAGE,
                                           "docker_builder_version": "1.1.2", "architecture": "linux/amd64"}))
        self.args = argparse.Namespace(profile=self.profile, prepared_dir=self.base / "prepared",
            record=self.base / "record.json", navigator=None, variant=["explicit"])
        self.calls = []

    def tearDown(self):
        self.directory.cleanup()

    def runner(self, responder=None):
        def call(argv, **kwargs):
            self.calls.append((list(map(str, argv)), kwargs))
            return responder(argv, kwargs) if responder else acceptance.Result(0)
        runner = acceptance.Runner(self.args, call)
        runner.temp = self.base / "private"
        runner.temp.mkdir(mode=0o700)
        runner.docker = "/tools/docker"
        runner.navigator = "/tools/ansible-navigator"
        return runner

    def test_missing_binary_is_a_failure_not_skip(self):
        runner = self.runner()
        with patch.object(acceptance.shutil, "which", return_value=None):
            with self.assertRaises(acceptance.Failed) as found:
                runner.binaries()
        self.assertEqual(found.exception.stage, "executable")
        self.assertIn("missing", found.exception.reason)

    def test_daemon_permission_missing_and_timeout_have_distinct_results(self):
        for response, expected in ((acceptance.Result(1, "permission denied CANARY"), "endpoint-permission-denied"),
                                   (acceptance.Result(1, "Cannot connect"), "endpoint-unavailable"),
                                   (acceptance.Result(-9, "", True), "timeout")):
            runner = self.runner(lambda argv, kwargs: acceptance.Result(0, "unix:///selected.sock")
                                 if "context" in argv else response)
            with self.assertRaises(acceptance.Failed) as found:
                runner.daemon()
            self.assertEqual(found.exception.reason, expected)
            runner.cleanup()
            self.assertTrue(all(kwargs["timeout"] <= 30 for _, kwargs in self.calls))
            self.assertNotIn("CANARY", json.dumps(runner.record))

    def test_context_environment_is_preserved_without_alternate_retry(self):
        runner = self.runner(lambda argv, kwargs: acceptance.Result(0, "unix:///named.sock")
                             if "context" in argv else acceptance.Result(0, '"29.8.2"'))
        runner.env.update(DOCKER_CONTEXT="selected", DOCKER_HOST="unix:///explicit.sock", DOCKER_TLS_VERIFY="1")
        runner.daemon()
        self.assertEqual(len(self.calls), 3)
        for _, kwargs in self.calls:
            self.assertEqual(kwargs["env"]["DOCKER_CONTEXT"], "selected")
            self.assertEqual(kwargs["env"]["DOCKER_HOST"], "unix:///explicit.sock")
        self.assertEqual(runner.record["docker_selection"]["source"], "DOCKER_CONTEXT")

    def test_image_unavailable_fails_without_pull(self):
        runner = self.runner(lambda argv, kwargs: acceptance.Result(1, "No such image"))
        with self.assertRaises(acceptance.Failed) as found:
            runner.images()
        self.assertEqual(found.exception.stage, "images")
        self.assertFalse(any("pull" in argv for argv, _ in self.calls))

    def test_startup_rejects_version_response_or_unrelated_event(self):
        for output, events in (("Docker version 29.8.2", event()),
                               (f"ACCEPTANCE_STARTUP:{NONCE}", event(run="unrelated"))):
            runner = self.runner(lambda argv, kwargs: acceptance.Result(0, output))
            with patch.object(acceptance.uuid, "uuid4") as generated:
                generated.return_value.hex = NONCE
                with patch.object(runner, "event_log", return_value=events):
                    with self.assertRaises(acceptance.Failed) as found:
                        runner.startup()
            self.assertEqual(found.exception.reason, "correlated-launch-proof-missing")
            runner.cleanup()

    def test_startup_records_only_correlated_container_and_actual_versions(self):
        versions = {"python": "3.12.1", "ansible_core": "2.17.2", "runner": "2.4.1", "ssh": "OpenSSH_9.9p1"}
        output = f"ACCEPTANCE_STARTUP:{NONCE}\nACCEPTANCE_VERSIONS:" + json.dumps(versions)
        runner = self.runner(lambda argv, kwargs: acceptance.Result(0, output))
        with patch.object(acceptance.uuid, "uuid4") as generated:
            generated.return_value.hex = NONCE
            with patch.object(runner, "event_log", return_value=event(run=runner.run_id)):
                runner.startup()
        stage = runner.record["stages"][-1]
        self.assertEqual(stage["ee_container_id"], CONTAINER)
        self.assertEqual(stage["versions"]["ssh"], "9.9p1")
        argv, call = self.calls[0]
        self.assertEqual(call["timeout"], 120)
        self.assertEqual(argv[1:3], ["exec", "--"])
        import shlex
        self.assertEqual(shlex.split(argv[3])[:2], ["python3", "-c"])
        settings = json.loads(Path(call["env"]["ANSIBLE_NAVIGATOR_CONFIG"]).read_text())
        self.assertEqual(settings["ansible-navigator"]["execution-environment"]["pull"], {"policy": "never"})
        self.assertIn("--tty=false", settings["ansible-navigator"]["execution-environment"]["container-options"])

    def test_source_install_matches_checksum_inside_isolated_plugin_tree(self):
        def response(argv, kwargs):
            if "build" in argv:
                Path(argv[argv.index("-o") + 1]).write_bytes(b"source binary")
            elif "install" in argv:
                folder = Path(kwargs["env"]["PACKER_PLUGIN_PATH"]) / acceptance.NAV_SOURCE
                folder.mkdir(parents=True)
                (folder / "packer-plugin-ansible-navigator_v0.1.0_x5.0_linux_amd64").write_bytes(b"source binary")
            return acceptance.Result(0)
        runner = self.runner(response)
        runner.go, runner.packer = "/tools/go", "/tools/packer"
        builder = self.args.prepared_dir / "plugins" / acceptance.BUILDER_SOURCE
        builder.mkdir(parents=True)
        (builder / "packer-plugin-docker_v1.1.2_x5.0_linux_amd64").write_bytes(b"prepared builder")
        normal = self.base / "normal-user-installation"
        normal.mkdir()
        sentinel = normal / "sentinel"
        sentinel.write_text("preserved")
        runner.env["PACKER_PLUGIN_PATH"] = str(normal)
        runner.install_source()
        self.assertEqual(sentinel.read_text(), "preserved")
        self.assertTrue(runner.installed_plugin.is_relative_to(runner.temp))
        self.assertEqual(acceptance.digest(runner.installed_plugin), runner.record["stages"][-1]["binary_sha256"])
        self.assertFalse(any("init" in argv for argv, _ in self.calls))
        for argv, kwargs in self.calls:
            self.assertEqual(kwargs["env"]["PACKER_PLUGIN_PATH"], str(runner.temp / "plugins"))

    def test_source_install_rejects_shadow_binary_with_different_checksum(self):
        def response(argv, kwargs):
            if "build" in argv:
                Path(argv[argv.index("-o") + 1]).write_bytes(b"source binary")
            elif "install" in argv:
                folder = Path(kwargs["env"]["PACKER_PLUGIN_PATH"]) / acceptance.NAV_SOURCE
                folder.mkdir(parents=True)
                (folder / "packer-plugin-ansible-navigator_v0.1.0_x5.0_linux_amd64").write_bytes(b"wrong binary")
            return acceptance.Result(0)
        runner = self.runner(response)
        runner.go, runner.packer = "/tools/go", "/tools/packer"
        builder = self.args.prepared_dir / "plugins" / acceptance.BUILDER_SOURCE
        builder.mkdir(parents=True)
        (builder / "packer-plugin-docker_v1.1.2_x5.0_linux_amd64").write_bytes(b"prepared builder")
        with self.assertRaises(acceptance.Failed) as found:
            runner.install_source()
        self.assertEqual(found.exception.reason, "installed-source-checksum-mismatch")

    def test_failed_build_retains_actual_selected_plugin_and_variant_identity(self):
        selected_path = self.base / "installed-source-plugin"
        selected_path.write_bytes(b"built source")
        def response(argv, kwargs):
            if "required" in argv:
                return acceptance.Result(0, str(selected_path))
            if "build" in argv:
                return acceptance.Result(1, "cannot attach stdin CANARY_PRIVATE_VALUE")
            return acceptance.Result(0)
        runner = self.runner(response)
        runner.packer = "/tools/packer"
        runner.installed_plugin = selected_path
        runner.source_plugin_checksum = acceptance.digest(selected_path)
        with patch.object(runner, "event_log", return_value=""):
            with self.assertRaises(acceptance.Failed):
                runner.build("explicit")
        variant = runner.record["variants"][0]
        self.assertEqual(variant["status"], "failed")
        self.assertEqual(variant["selected_plugin_path"], str(selected_path))
        self.assertEqual(variant["plugin_sha256"], acceptance.digest(selected_path))
        self.assertIn(variant["build_id"], runner.owned)
        self.assertTrue(variant["nonce"])
        self.assertEqual(variant["packer_exit"], 1)
        stages = runner.record["stages"]
        self.assertEqual([stage["name"] for stage in stages], ["validation", "source-plugin-selection"])
        self.assertEqual(stages[1]["selected_path"], str(selected_path))
        self.assertNotIn("CANARY", json.dumps(runner.record))

    def test_validation_failure_keeps_failed_variant_without_claiming_plugin_selection(self):
        runner = self.runner(lambda argv, kwargs: acceptance.Result(1, "CANARY_SECRET"))
        runner.packer = "/tools/packer"
        with self.assertRaises(acceptance.Failed):
            runner.build("explicit")
        variant = runner.record["variants"][0]
        self.assertEqual(variant["failed_stage"], "validation")
        self.assertEqual(variant["status"], "failed")
        self.assertNotIn("selected_plugin_path", variant)
        self.assertFalse(runner.record["stages"])
        self.assertNotIn("CANARY", json.dumps(runner.record))

    def test_selection_rejects_path_or_checksum_drift_before_build(self):
        selected_path = self.base / "installed-source-plugin"
        selected_path.write_bytes(b"changed binary")
        runner = self.runner(lambda argv, kwargs: acceptance.Result(0, str(selected_path)))
        runner.packer = "/tools/packer"
        runner.installed_plugin = selected_path
        runner.source_plugin_checksum = "previous-checksum"
        with self.assertRaises(acceptance.Failed) as found:
            runner.build("explicit")
        self.assertEqual(found.exception.stage, "source-plugin")
        self.assertFalse(any("build" in argv for argv, _ in self.calls))
        self.assertFalse(any(stage["name"] == "source-plugin-selection" for stage in runner.record["stages"]))

    def test_startup_classifies_missing_image_error_and_timeout(self):
        for response, expected in ((acceptance.Result(1, "No such image"), "image-unavailable"),
                                   (acceptance.Result(1, "Runner startup broke"), "ee-startup-failed"),
                                   (acceptance.Result(-9, "", True), "timeout")):
            runner = self.runner(lambda argv, kwargs: response)
            with self.assertRaises(acceptance.Failed) as found:
                runner.startup()
            self.assertEqual(found.exception.reason, expected)
            self.assertEqual(self.calls[-1][1]["timeout"], 120)
            runner.docker = None
            runner.cleanup()

    def test_cleanup_removes_only_labeled_ids_never_images(self):
        queries = 0
        def response(argv, kwargs):
            nonlocal queries
            if "ps" in argv:
                queries += 1
                return acceptance.Result(0, CONTAINER if queries == 1 else "")
            return acceptance.Result(0)
        runner = self.runner(response)
        owned = runner.run_id
        runner.owned.append(owned)
        self.assertTrue(runner.cleanup())
        self.assertFalse((self.base / "private").exists())
        self.assertEqual(self.calls[0][0][-1], f"label={acceptance.LABEL}={owned}")
        self.assertEqual(self.calls[1][0], ["/tools/docker", "rm", "-f", CONTAINER])
        self.assertFalse(any("prune" in argv or "rmi" in argv for argv, _ in self.calls))

    def test_incomplete_cleanup_prevents_full_pass(self):
        runner = self.runner(lambda argv, kwargs: acceptance.Result(1, "CANARY"))
        runner.owned.append(runner.run_id)
        self.assertFalse(runner.cleanup())
        self.assertEqual(runner.record["cleanup"]["status"], "failed")
        self.assertNotIn("CANARY", json.dumps(runner.record))

    def test_explicit_acceptance_failure_still_writes_sanitized_record_and_cleans(self):
        runner = self.runner()
        old_temp = runner.temp
        old_temp.rmdir()
        runner.docker = None
        with patch.object(runner, "binaries", side_effect=acceptance.Failed("daemon", "endpoint-unavailable",
                  acceptance.Result(1, "PRIVATE_KEY=CANARY"))):
            self.assertEqual(runner.accept(), 1)
        record = json.loads(self.args.record.read_text())
        self.assertEqual(record["result"], "failed")
        self.assertEqual(record["cleanup"]["status"], "passed")
        self.assertNotIn("CANARY", self.args.record.read_text())

    def test_interruption_cleans_owned_state_and_cannot_pass(self):
        runner = self.runner()
        runner.temp.rmdir()
        runner.owned.append(runner.run_id)
        with patch.object(runner, "binaries", side_effect=KeyboardInterrupt):
            self.assertEqual(runner.accept(), 1)
        record = json.loads(self.args.record.read_text())
        self.assertEqual(record["cleanup"]["status"], "passed")
        self.assertEqual(record["result"], "failed")
        self.assertEqual(record["stages"][-1]["reason"], "interrupted")
        self.assertTrue(self.calls)
        self.assertTrue(all("ps" in argv for argv, _ in self.calls))

    def test_generated_temporary_files_are_cleaned_outside_work_mount(self):
        runner = self.runner()
        runner.temp.rmdir()
        private_paths = []
        def interrupt():
            temp_path = Path(runner.env["TMPDIR"])
            private_paths.append(temp_path)
            self.assertFalse(temp_path.is_relative_to(runner.work))
            (temp_path / "generated-private-key").write_text("CANARY")
            raise KeyboardInterrupt
        runner.docker = None
        with patch.object(runner, "binaries", side_effect=interrupt):
            self.assertEqual(runner.accept(), 1)
        self.assertFalse(private_paths[0].exists())
        self.assertNotIn("CANARY", self.args.record.read_text())

    def test_mutable_image_is_rejected(self):
        self.profile.write_text(json.dumps({"target_image": "python:latest"}))
        with self.assertRaises(acceptance.Failed):
            acceptance.load_profile(self.profile)

    def test_event_collection_has_fixed_until_and_timeout(self):
        runner = self.runner(lambda argv, kwargs: acceptance.Result(-9, "", True))
        with self.assertRaises(acceptance.Failed) as found:
            runner.event_log(RUN, "1234567890")
        self.assertEqual(found.exception.reason, "timeout")
        self.assertIn("--until", self.calls[-1][0])
        self.assertEqual(self.calls[-1][1]["timeout"], 10)

    def test_variant_probes_use_effective_navigator_path_and_same_docker_environment(self):
        runner = self.runner()
        runner.navigator_selected = runner.navigator
        fake_env = dict(runner.env, PATH="/restricted", DOCKER_CONTEXT="preserved-context")
        with patch.object(runner, "variant_env", return_value=(fake_env, "ansible-navigator", ["/configured"])), \
             patch.object(acceptance, "executable", side_effect=lambda command, env: "/configured/" + command), \
             patch.object(acceptance, "real_navigator", side_effect=lambda path, env, run: path), \
             patch.object(runner, "daemon") as daemon, patch.object(runner, "startup") as startup:
            runner.probe_variant("configured-path")
        docker_env, docker_path, variant = daemon.call_args.args
        navigator_env, navigator_path, startup_docker, startup_variant = startup.call_args.args
        self.assertEqual(docker_env["PATH"], "/configured:/restricted")
        self.assertEqual(docker_env["DOCKER_CONTEXT"], "preserved-context")
        self.assertEqual(navigator_env, docker_env)
        self.assertEqual(navigator_path, "/configured/ansible-navigator")
        self.assertEqual(startup_docker, docker_path)
        self.assertEqual(variant, startup_variant)
        self.assertEqual(fake_env["PATH"], "/restricted", "must not mutate Packer's inherited PATH")

    def test_configured_path_is_absent_from_inherited_path(self):
        runner = self.runner()
        runner.navigator_selected = runner.navigator
        env, command, search = runner.variant_env("configured-path")
        self.assertEqual(command, "ansible-navigator")
        self.assertEqual(search, ["/tools"])
        self.assertIsNone(acceptance.shutil.which(command, path=env["PATH"]))
        self.assertNotEqual(env["PATH"], runner.env.get("PATH"))


class ProcessTests(unittest.TestCase):
    def test_timeout_kills_spawned_process_group(self):
        with tempfile.TemporaryDirectory() as directory:
            delayed = Path(directory) / "survived"
            descendant = f"import time;from pathlib import Path;time.sleep(0.4);Path({str(delayed)!r}).touch()"
            payload = ("import subprocess,sys,time;"
                       f"subprocess.Popen([sys.executable,'-c',{descendant!r}]);time.sleep(60)")
            started = time.monotonic()
            result = acceptance.execute([sys.executable, "-c", payload], timeout=0.1)
            self.assertTrue(result.timed_out)
            self.assertLess(time.monotonic() - started, 2)
            time.sleep(0.5)
            self.assertFalse(delayed.exists(), "descendant survived the stage timeout")


if __name__ == "__main__":
    unittest.main()
