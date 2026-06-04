import subprocess
import unittest
from pathlib import Path
import sys


TOOLS_DIR = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOLS_DIR))

from crack_build_mode import determine_build_mode


class BuildCrackedScriptTests(unittest.TestCase):
    def test_pipeline_and_account_nurturing_image_updates_are_go_only(self):
        result = determine_build_mode(
            [
                "assets/resource/pipeline/AccountNurturing/AccountNurturing.json",
                "assets/resource/pipeline/AccountNurturing/AccountNurturingSynchroDeviceEnhance.json",
                "assets/resource/image/AccountNurturing/ConsumablesOFF.png",
            ]
        )

        self.assertEqual(result.mode, "go-only")
        self.assertEqual(result.script_args, ["-Yes"])
        self.assertIn("no full-rebuild rule matched", result.reason)

    def test_full_when_deps_change(self):
        result = determine_build_mode(["deps/bin/MaaFramework.dll"])

        self.assertEqual(result.mode, "full")
        self.assertEqual(result.script_args, ["-Full", "-Yes"])
        self.assertIn("deps/", result.reason)

    def test_full_when_non_membership_go_source_changes(self):
        result = determine_build_mode(["agent/go-service/pkg/resource/resource_sink.go"])

        self.assertEqual(result.mode, "full")
        self.assertIn("agent/go-service", result.reason)

    def test_membership_go_source_does_not_force_full_by_itself(self):
        result = determine_build_mode(
            ["agent/go-service/taskersink/membership/memberdata.go"]
        )

        self.assertEqual(result.mode, "go-only")

    def test_full_when_install_script_changes(self):
        result = determine_build_mode(["tools/install.py"])

        self.assertEqual(result.mode, "full")

    def test_full_when_assets_interface_changes(self):
        result = determine_build_mode(["assets/interface.json"])

        self.assertEqual(result.mode, "full")

    def test_full_when_assets_tasks_change(self):
        result = determine_build_mode(["assets/tasks/AccountNurturing.json"])

        self.assertEqual(result.mode, "full")

    def test_docs_scripts_and_gitignore_are_go_only(self):
        result = determine_build_mode(
            [
                ".gitignore",
                "README.md",
                ".github/workflows/mirrorchyan_release.yml",
                "tools/configure.py",
                "maatools.config.mts",
            ]
        )

        self.assertEqual(result.mode, "go-only")

    def test_cli_reads_git_diff_name_only_text(self):
        repo_root = Path(__file__).resolve().parents[2]
        diff_text = "\n".join(
            [
                "assets/resource/pipeline/AccountNurturing/AccountNurturing.json",
                "assets/resource/image/AccountNurturing/HollowDot.png",
            ]
        )

        result = subprocess.run(
            ["python", "tools\\crack_build_mode.py", "--format", "args"],
            cwd=repo_root,
            input=diff_text,
            capture_output=True,
            text=True,
            encoding="utf-8",
            check=False,
        )

        self.assertEqual(result.returncode, 0, msg=result.stderr)
        self.assertEqual(result.stdout.strip(), "-Yes")

    def test_build_cracked_reports_bypass_marker_present(self):
        repo_root = Path(__file__).resolve().parents[2]
        result = subprocess.run(
            ["pwsh", "tools\\build-cracked.ps1", "-Yes"],
            cwd=repo_root,
            capture_output=True,
            text=True,
            encoding="utf-8",
            errors="replace",
            check=False,
        )

        self.assertEqual(result.returncode, 0, msg=result.stdout + "\n" + result.stderr)
        self.assertIn("bypass-log string present in binary", result.stdout)
        self.assertNotIn("bypass-log string NOT found", result.stdout)


if __name__ == "__main__":
    unittest.main()
