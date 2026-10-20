#!/usr/bin/env python3
"""Tests for the sharded suite runner."""
import argparse
import contextlib
import importlib.util
import io
import json
from pathlib import Path
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
