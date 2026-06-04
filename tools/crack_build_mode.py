import argparse
import json
import sys
from dataclasses import dataclass
from pathlib import PurePosixPath
from typing import Iterable


@dataclass(frozen=True)
class BuildModeDecision:
    mode: str
    reason: str

    @property
    def script_args(self) -> list[str]:
        if self.mode == "full":
            return ["-Full", "-Yes"]
        return ["-Yes"]


def _normalize_path(path: str) -> str:
    return path.strip().replace("\\", "/").lstrip("./")


def _is_under(path: str, prefix: str) -> bool:
    return path == prefix.rstrip("/") or path.startswith(prefix.rstrip("/") + "/")


def _first_full_reason(path: str) -> str | None:
    if _is_under(path, "deps"):
        return f"{path}: deps/ changed"

    if path == "tools/install.py":
        return f"{path}: tools/install.py changed"

    if path == "assets/interface.json":
        return f"{path}: assets/interface.json changed"

    if _is_under(path, "assets/tasks"):
        return f"{path}: assets/tasks changed"

    if (
        _is_under(path, "agent/go-service")
        and PurePosixPath(path).suffix == ".go"
        and not _is_under(path, "agent/go-service/taskersink/membership")
    ):
        return f"{path}: non-membership agent/go-service Go source changed"

    return None


def determine_build_mode(changed_files: Iterable[str]) -> BuildModeDecision:
    paths = [_normalize_path(p) for p in changed_files if _normalize_path(p)]

    for path in paths:
        reason = _first_full_reason(path)
        if reason:
            return BuildModeDecision(mode="full", reason=reason)

    return BuildModeDecision(
        mode="go-only",
        reason="no full-rebuild rule matched",
    )


def _read_changed_files(file_path: str | None) -> list[str]:
    if file_path:
        with open(file_path, "r", encoding="utf-8") as f:
            return f.read().splitlines()
    return sys.stdin.read().splitlines()


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Choose MDA cracked build mode from git diff --name-only output."
    )
    parser.add_argument(
        "file",
        nargs="?",
        help="Optional text file containing git diff --name-only output. Reads stdin by default.",
    )
    parser.add_argument(
        "--format",
        choices=("text", "args", "json"),
        default="text",
        help="Output format. args emits arguments for tools\\build-cracked.ps1.",
    )
    args = parser.parse_args()

    decision = determine_build_mode(_read_changed_files(args.file))

    if args.format == "args":
        print(" ".join(decision.script_args))
    elif args.format == "json":
        print(
            json.dumps(
                {
                    "mode": decision.mode,
                    "script_args": decision.script_args,
                    "reason": decision.reason,
                },
                ensure_ascii=False,
            )
        )
    else:
        print(f"mode={decision.mode}")
        print(f"args={' '.join(decision.script_args)}")
        print(f"reason={decision.reason}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
