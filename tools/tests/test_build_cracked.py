import subprocess
import unittest
from pathlib import Path


class BuildCrackedScriptTests(unittest.TestCase):
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
