#!/usr/bin/env python3
"""The suite summary must preserve the assertion that made a subtest fail."""
import contextlib
import importlib.util
import io
import json
from pathlib import Path
import tempfile
import unittest

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

if __name__ == "__main__":
    unittest.main()
