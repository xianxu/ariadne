#!/usr/bin/env python3
"""Deterministic deadline and process ownership tests for the cost runner."""

import importlib.util
from pathlib import Path
import signal
import subprocess
import unittest
from unittest import mock

spec = importlib.util.spec_from_file_location(
    "landing_drop_cost", Path(__file__).with_name("test-landing-drop-cost.py")
)
cost = importlib.util.module_from_spec(spec)
spec.loader.exec_module(cost)


class DeadlineTest(unittest.TestCase):
    def test_overall_budget_uses_one_monotonic_deadline(self):
        with mock.patch.object(cost.time, "monotonic", return_value=20):
            self.assertEqual(cost.execution_deadline(10, None), 30)
        with mock.patch.object(cost.time, "monotonic", return_value=25):
            self.assertEqual(cost.remaining(30), 5)

    def test_inherited_deadline_accounts_for_script_startup(self):
        with mock.patch.object(cost.time, "monotonic", return_value=20), \
             mock.patch.object(cost.time, "time", return_value=40):
            self.assertEqual(cost.execution_deadline(None, 45_000_000_000), 25)

    def test_exhausted_budget_refuses_before_launch(self):
        with mock.patch.object(cost.time, "monotonic", return_value=30), \
             mock.patch.object(cost.subprocess, "Popen") as launch:
            with self.assertRaises(TimeoutError):
                cost.run_command(["unused"], Path.cwd(), 30, supervised=True)
        launch.assert_not_called()

    def test_manual_budget_must_be_positive_and_finite(self):
        for budget in (0, -1, float("nan"), float("inf")):
            with self.subTest(budget=budget), self.assertRaises(ValueError):
                cost.execution_deadline(budget, None)


class ProcessBudgetTest(unittest.TestCase):
    def child(self):
        child = mock.Mock(pid=123, returncode=0)
        child.communicate.return_value = ("output", "")
        return child

    def test_supervised_child_inherits_go_owned_group_and_remaining_budget(self):
        child = self.child()
        with mock.patch.object(cost.time, "monotonic", return_value=20), \
             mock.patch.object(cost.subprocess, "Popen", return_value=child) as launch:
            result = cost.run_command(["tool"], Path.cwd(), 25, supervised=True)
        self.assertFalse(launch.call_args.kwargs["start_new_session"])
        child.communicate.assert_called_once_with(timeout=5)
        self.assertEqual(result.stdout, "output")

    def test_supervised_timeout_leaves_whole_group_to_go_authority(self):
        child = self.child()
        child.communicate.side_effect = subprocess.TimeoutExpired(["tool"], 5)
        with mock.patch.object(cost.time, "monotonic", return_value=20), \
             mock.patch.object(cost.subprocess, "Popen", return_value=child), \
             mock.patch.object(cost.os, "killpg") as kill:
            with self.assertRaises(subprocess.TimeoutExpired):
                cost.run_command(["tool"], Path.cwd(), 25, supervised=True)
        kill.assert_not_called()
        child.kill.assert_not_called()

    def test_manual_timeout_kills_and_joins_the_entire_owned_child_group(self):
        child = self.child()
        child.communicate.side_effect = [subprocess.TimeoutExpired(["tool"], 5), ("partial", "error")]
        with mock.patch.object(cost.time, "monotonic", return_value=20), \
             mock.patch.object(cost.subprocess, "Popen", return_value=child) as launch, \
             mock.patch.object(cost.os, "killpg") as kill:
            with self.assertRaises(subprocess.TimeoutExpired):
                cost.run_command(["tool"], Path.cwd(), 25, supervised=False)
        self.assertTrue(launch.call_args.kwargs["start_new_session"])
        kill.assert_called_once_with(123, signal.SIGKILL)
        self.assertEqual(child.communicate.call_count, 2)


if __name__ == "__main__":
    unittest.main()
