#!/usr/bin/env python3
"""Tests for the sharded suite runner."""
import argparse
import contextlib
import importlib.util
import io
import json
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest
from unittest import mock

spec = importlib.util.spec_from_file_location("test_shard", Path(__file__).with_name("test-shard.py"))
shard = importlib.util.module_from_spec(spec)
spec.loader.exec_module(shard)

class ReportTest(unittest.TestCase):
    def test_failed_subtest_keeps_its_assertion(self):
        package = "example/weavefs"
        events = [
            {"Action": "output", "Package": package, "Test": "TestLease/owned=true", "Output": "lease_test.go:42: descriptor retained\\n"},
            {"Action": "fail", "Package": package, "Test": "TestLease/owned=true"},
            {"Action": "fail", "Package": package, "Test": "TestLease"},
            {"Action": "fail", "Package": package},
        ]
        with tempfile.TemporaryDirectory(prefix="shard-report-") as root:
            path = Path(root)
            (path / "packages.json").write_text("".join(json.dumps(x) + "\n" for x in events))
            out = io.StringIO()
            with contextlib.redirect_stdout(out):
                failed, _, _ = shard.report(path, [("packages", 1, 0.1)])
        self.assertTrue(failed)
        self.assertIn("TestLease/owned=true", out.getvalue())
        self.assertIn("descriptor retained", out.getvalue())

    def report_of(self, events, code=1):
        with tempfile.TemporaryDirectory(prefix="shard-report-") as root:
            path = Path(root)
            (path / "shard3.json").write_text("".join(json.dumps(x) + "\n" for x in events))
            out = io.StringIO()
            with contextlib.redirect_stdout(out):
                failed, _, _ = shard.report(path, [("shard3", code, 1800.4)])
        return failed, out.getvalue()

    def test_alarm_prints_a_marker_naming_the_running_tests(self):
        package = "example/sdlc"
        alarm = "panic: test timed out after 30m0s\n\trunning tests:\n\t\tTestSlow (29m59s)\n\t\tTestSlow/sub (2m0s)\n\t\tTestOther (1m2s)\n\ngoroutine 1 [running]:\n"
        events = [
            {"Action": "output", "Package": package, "Test": "TestSlow", "Output": "=== RUN   TestSlow\n"},
            {"Action": "output", "Package": package, "Output": alarm},
            {"Action": "fail", "Package": package},
        ]
        failed, out = self.report_of(events)
        self.assertTrue(failed)
        self.assertIn("--- shard3 hit its 30m0s test alarm in example/sdlc; running: TestSlow, TestSlow/sub, TestOther\n", out)

    def test_each_alarmed_package_gets_a_marker(self):
        def alarm(pkg):
            return [{"Action": "output", "Package": pkg, "Output": "panic: test timed out after 5m0s\n"},
                    {"Action": "output", "Package": pkg, "Output": "\trunning tests:\n"},
                    {"Action": "output", "Package": pkg, "Output": "\t\tTestHang (5m0s)\n"},
                    {"Action": "fail", "Package": pkg}]
        _, out = self.report_of(alarm("example/a") + [{"Action": "pass", "Package": "example/b"}] + alarm("example/c"))
        self.assertIn("hit its 5m0s test alarm in example/a; running: TestHang", out)
        self.assertIn("hit its 5m0s test alarm in example/c; running: TestHang", out)
        self.assertNotIn("example/b;", out)

    def test_no_alarm_no_marker(self):
        package = "example/sdlc"
        events = [
            {"Action": "output", "Package": package, "Test": "TestX", "Output": "x_test.go:9: boom\n"},
            {"Action": "fail", "Package": package, "Test": "TestX"},
            {"Action": "fail", "Package": package},
        ]
        _, out = self.report_of(events)
        self.assertNotIn("test alarm", out)

    @unittest.skipUnless(shutil.which("go"), "needs the go toolchain")
    def test_real_go_alarm_through_both_runner_paths(self):
        # The live check behind the hand-written fixtures above (#630): the real
        # toolchain's alarm, through go test -json (the packages job) and
        # through test2json over a compiled binary (a shard), yields the marker.
        with tempfile.TemporaryDirectory(prefix="shard-alarm-") as root:
            mod = Path(root) / "mod"
            mod.mkdir()
            (mod / "go.mod").write_text("module example.com/alarm\n\ngo 1.21\n")
            (mod / "a_test.go").write_text('package alarm\n\nimport (\n\t"testing"\n\t"time"\n)\n\nfunc TestHang(t *testing.T) { time.Sleep(10 * time.Second) }\n')
            out = Path(root) / "out"
            out.mkdir()
            with open(out / "packages.json", "w") as f:
                subprocess.run(["go", "test", "-json", "-count=1", "-timeout=1s", "."], cwd=mod, stdout=f, stderr=subprocess.STDOUT)
            binary = Path(root) / "alarm.test"
            subprocess.run(["go", "test", "-c", "-o", str(binary), "."], cwd=mod, check=True)
            with open(out / "shard0.json", "w") as f:
                subprocess.run(["go", "tool", "test2json", "-t", "-p", "example.com/alarm", str(binary), "-test.v=test2json", "-test.timeout=1s"],
                               cwd=mod, stdout=f, stderr=subprocess.STDOUT)
            printed = io.StringIO()
            with contextlib.redirect_stdout(printed):
                failed, _, _ = shard.report(out, [("packages", 1, 1.0), ("shard0", 1, 1.0)])
        self.assertTrue(failed)
        for label in ("packages", "shard0"):
            self.assertIn(f"--- {label} hit its 1s test alarm in example.com/alarm; running: TestHang", printed.getvalue())


class PackageSelectionTest(unittest.TestCase):
    OTHERS = ["example/a", "example/b", "example/c"]

    def test_unset_or_empty_runs_everything(self):
        for listed in (None, "", "  "):
            self.assertEqual(shard.select_packages(listed, self.OTHERS), (True, self.OTHERS))

    def test_a_listed_subset_runs_only_those_packages(self):
        run_main, others = shard.select_packages("example/b", self.OTHERS)
        self.assertEqual((run_main, others), (False, ["example/b"]))
        jobs = shard.build_jobs(others, [], Path("sdlc.test"), Path("cmd/sdlc"), Path("."), "30m")
        self.assertEqual([label for label, _, _ in jobs], ["packages"])
        self.assertEqual(jobs[0][1][-1:], ["example/b"])
        self.assertNotIn("example/a", jobs[0][1])

    def test_listing_cmd_sdlc_runs_its_shards_alone(self):
        run_main, others = shard.select_packages(shard.MAIN_PKG, self.OTHERS)
        self.assertEqual((run_main, others), (True, []))
        jobs = shard.build_jobs(others, [["TestA"], ["TestB"]], Path("sdlc.test"), Path("cmd/sdlc"), Path("."), "30m")
        self.assertEqual([label for label, _, _ in jobs], ["shard0", "shard1"])

    def test_an_unknown_package_refuses(self):
        with self.assertRaises(ValueError):
            shard.select_packages("example/a example/zz", self.OTHERS)

    def test_a_selection_with_no_jobs_exits_cleanly(self):
        # cmd/sdlc listed alone but compiling no tests leaves no job; the pool still needs a worker.
        with tempfile.TemporaryDirectory() as root, \
             mock.patch.object(shard, "git", return_value=root), \
             mock.patch.object(shard, "git_list", return_value=[shard.MAIN_PKG]), \
             mock.patch.object(shard.subprocess, "run"), \
             mock.patch.object(shard.subprocess, "check_output", return_value=""), \
             mock.patch.object(shard.sys, "argv", ["test-shard.py"]), \
             mock.patch.dict(shard.os.environ, {"SDLC_AFFECTED_GO_PKGS": shard.MAIN_PKG}):
            self.assertEqual(shard.main(), 0)


class DefaultShardTest(unittest.TestCase):
    def parse_main_args(self, argv, cpu_count=32):
        class Parsed(Exception):
            pass

        captured = {}
        original = argparse.ArgumentParser.parse_args

        def capture(parser, args=None, namespace=None):
            captured["args"] = original(parser, args, namespace)
            raise Parsed

        with mock.patch.object(shard.os, "cpu_count", return_value=cpu_count), \
             mock.patch.object(shard.sys, "argv", ["test-shard.py", *argv]), \
             mock.patch.object(argparse.ArgumentParser, "parse_args", capture):
            with self.assertRaises(Parsed):
                shard.main()
        return captured["args"]

    def test_default_shards_caps_large_cpu_counts_at_eight(self):
        for cpus in (8, 12, 32):
            with self.subTest(cpus=cpus):
                self.assertEqual(shard.default_shards(cpus), 8)

    def test_default_shards_uses_positive_small_cpu_count(self):
        self.assertEqual(shard.default_shards(1), 1)
        self.assertEqual(shard.default_shards(3), 3)

    def test_default_shards_falls_back_to_four_when_unavailable(self):
        for cpus in (None, 0):
            with self.subTest(cpus=cpus):
                self.assertEqual(shard.default_shards(cpus), 4)

    def test_cli_default_uses_capped_detected_cpu_count(self):
        self.assertEqual(self.parse_main_args([]).shards, 8)

    def test_cli_explicit_shards_override_default(self):
        self.assertEqual(self.parse_main_args(["--shards", "12"]).shards, 12)

    def test_makefile_reserves_both_slots_before_workflow_admission(self):
        makefile = Path(__file__).parents[1] / "Makefile"
        lines = makefile.read_text().splitlines()
        weight_index = next(i for i, line in enumerate(lines) if line.startswith("WF_TEST_WEIGHT ?="))
        self.assertEqual(lines[weight_index], "WF_TEST_WEIGHT ?= 2")
        workflow_include = lines.index("-include Makefile.workflow")
        self.assertLess(weight_index, workflow_include)

if __name__ == "__main__":
    unittest.main()
