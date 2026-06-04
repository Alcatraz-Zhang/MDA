import importlib.util
import json
import sys
import tempfile
import types
import unittest
from pathlib import Path
from unittest import mock


EXPECTED_EXPOSURES = (
    {
        "import": "tasks/AccountNurturing.json",
        "task_file": "AccountNurturing.json",
        "task": "AccountNurturing",
        "options": (
            "AccountNurturingCharacterBreakthrough",
            "AccountNurturingSynchroDeviceEnhance",
        ),
    },
    {
        "import": "tasks/RedDotClear.json",
        "task_file": "RedDotClear.json",
        "task": "RedDotClear",
        "options": ("RedDotClearArchives",),
    },
)


def load_install_module():
    install_path = Path(__file__).resolve().parents[1] / "install.py"
    module_name = "install_under_task_exposure_smoke_test"
    spec = importlib.util.spec_from_file_location(module_name, install_path)
    module = importlib.util.module_from_spec(spec)
    original_argv = sys.argv[:]
    original_sys_path = sys.path[:]
    original_jsonc = sys.modules.get("jsonc")
    sys.argv = [str(install_path), "v-smoke", "win", "x86_64"]
    sys.path.insert(0, str(install_path.parent))
    if original_jsonc is None:
        jsonc = types.ModuleType("jsonc")
        jsonc.load = json.load
        jsonc.dump = json.dump
        sys.modules["jsonc"] = jsonc
    try:
        spec.loader.exec_module(module)
    finally:
        sys.argv = original_argv
        sys.path[:] = original_sys_path
        if original_jsonc is None:
            sys.modules.pop("jsonc", None)
        else:
            sys.modules["jsonc"] = original_jsonc
    return module


def read_json(path):
    with open(path, "r", encoding="utf-8") as f:
        return json.load(f)


def assert_expected_exposures(test_case, interface_path, tasks_dir):
    interface = read_json(interface_path)
    imports = set(interface["import"])

    for expected in EXPECTED_EXPOSURES:
        test_case.assertIn(expected["import"], imports)

        task_config = read_json(tasks_dir / expected["task_file"])
        task_names = {task["name"] for task in task_config["task"]}
        test_case.assertIn(expected["task"], task_names)

        task = next(task for task in task_config["task"] if task["name"] == expected["task"])
        task_options = set(task.get("option", ()))
        option_configs = task_config.get("option", {})

        for option_name in expected["options"]:
            test_case.assertIn(option_name, task_options)
            test_case.assertIn(option_name, option_configs)


class TaskExposureSmokeTests(unittest.TestCase):
    def test_source_assets_expose_new_upstream_tasks(self):
        repo_root = Path(__file__).resolve().parents[2]

        assert_expected_exposures(
            self,
            repo_root / "assets" / "interface.json",
            repo_root / "assets" / "tasks",
        )

    def test_install_resource_exposes_new_upstream_tasks(self):
        install = load_install_module()
        original_copytree = install.shutil.copytree
        original_copy2 = install.shutil.copy2

        def copy_runtime_subset(src, dst, *args, **kwargs):
            src = Path(src)
            dst = Path(dst)
            if src.name == "tasks":
                return original_copytree(src, dst, *args, **kwargs)
            dst.mkdir(parents=True, exist_ok=True)
            return dst

        with tempfile.TemporaryDirectory() as tmp_dir:
            install.install_path = Path(tmp_dir)
            install.version = "v-smoke"

            with (
                mock.patch.object(install, "configure_ocr_model", return_value=True),
                mock.patch.object(install.shutil, "copytree", side_effect=copy_runtime_subset),
                mock.patch.object(install.shutil, "copy2", side_effect=original_copy2),
            ):
                install.install_resource()

            assert_expected_exposures(
                self,
                install.install_path / "interface.json",
                install.install_path / "tasks",
            )


if __name__ == "__main__":
    unittest.main()
