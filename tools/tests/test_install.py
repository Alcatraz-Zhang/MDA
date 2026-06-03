import importlib.util
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock


def load_install_module():
    install_path = Path(__file__).resolve().parents[1] / "install.py"
    module_name = "install_under_test"
    spec = importlib.util.spec_from_file_location(module_name, install_path)
    module = importlib.util.module_from_spec(spec)
    original_argv = sys.argv[:]
    original_sys_path = sys.path[:]
    sys.argv = [str(install_path), "v0.0.1", "win", "x86_64"]
    sys.path.insert(0, str(install_path.parent))
    try:
        spec.loader.exec_module(module)
    finally:
        sys.argv = original_argv
        sys.path[:] = original_sys_path
    return module


class InstallAgentTests(unittest.TestCase):
    def test_force_rebuild_recompiles_existing_binary(self):
        install = load_install_module()

        with tempfile.TemporaryDirectory() as tmp_dir:
            install.install_path = Path(tmp_dir)
            install.os_name = "win"
            binary = install.install_path / "agent" / "go-service.exe"
            binary.parent.mkdir(parents=True, exist_ok=True)
            binary.write_bytes(b"old-binary")

            with mock.patch.object(install, "build_go_agent", return_value=True) as build_mock:
                install.install_agent(force_rebuild=True)

            build_mock.assert_called_once_with()


if __name__ == "__main__":
    unittest.main()
