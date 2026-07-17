import argparse
import os
import platform
import shutil
import stat
import subprocess
import sys
from pathlib import Path

from cli_support import Console, init_localization

PROJECT_BASE: Path = Path(__file__).parent.parent.resolve()

LOCALS_DIR = Path(__file__).parent / "locals" / "build_and_install"


_local_t = lambda key, **kwargs: key.format(**kwargs) if kwargs else key


def init_local() -> None:
    global _local_t
    t_func, load_error_path = init_localization(LOCALS_DIR)
    _local_t = t_func
    if load_error_path:
        print(Console.err(t("error_load_locale", path=load_error_path)))


def t(key: str, **kwargs) -> str:
    return _local_t(key, **kwargs)


def is_directory_link(path: Path) -> bool:
    """Return True for directory symlinks and Windows junction/reparse points."""
    if path.is_symlink():
        return True

    if platform.system() != "Windows":
        return False

    if hasattr(path, "is_junction") and path.is_junction():
        return True

    try:
        attrs = path.lstat().st_file_attributes
    except (AttributeError, FileNotFoundError, OSError):
        return False

    reparse_point = getattr(stat, "FILE_ATTRIBUTE_REPARSE_POINT", 0x400)
    return bool(attrs & reparse_point)


def remove_path_for_replacement(path: Path) -> None:
    """Remove a path before replacing it, without recursing into directory links."""
    if path.is_symlink():
        path.unlink(missing_ok=True)
    elif is_directory_link(path):
        path.rmdir()
    elif path.is_dir():
        shutil.rmtree(path)
    elif path.exists():
        path.unlink()


def create_directory_link(src: Path, dst: Path) -> bool:
    """
    在指定位置创建一个指定目录的链接
    - Windows：Junction
    - Unix/macOS：symlink
    """
    if dst.exists() or dst.is_symlink() or is_directory_link(dst):
        remove_path_for_replacement(dst)

    dst.parent.mkdir(parents=True, exist_ok=True)

    if platform.system() == "Windows":
        result = subprocess.run(
            ["cmd", "/c", "mklink", "/J", str(dst), str(src)],
            capture_output=True,
            text=True,
        )
        if result.returncode != 0:
            print(
                f"  {Console.err(t('error'))} {t('create_junction_failed')}: {result.stderr}"
            )
            return False
    else:
        dst.symlink_to(src)

    return True


def create_file_link(src: Path, dst: Path) -> bool:
    """创建文件链接（硬链接优先）"""
    if dst.exists() or dst.is_symlink():
        dst.unlink(missing_ok=True)

    dst.parent.mkdir(parents=True, exist_ok=True)

    if platform.system() == "Windows":
        result = subprocess.run(
            ["cmd", "/c", "mklink", "/H", str(dst), str(src)],
            capture_output=True,
            text=True,
        )
        if result.returncode != 0:
            result = subprocess.run(
                ["cmd", "/c", "mklink", str(dst), str(src)],
                capture_output=True,
                text=True,
            )
            if result.returncode != 0:
                print(
                    f"  {Console.err(t('error'))} {t('create_file_link_failed')}: {result.stderr}"
                )
                return False
    else:
        try:
            dst.hardlink_to(src)
        except (OSError, NotImplementedError):
            dst.symlink_to(src)

    return True


def copy_directory(src: Path, dst: Path) -> bool:
    """复制目录（替换）"""
    if dst.exists() or dst.is_symlink() or is_directory_link(dst):
        remove_path_for_replacement(dst)
    shutil.copytree(src, dst)
    return True


def copy_file(src: Path, dst: Path) -> bool:
    """复制文件"""
    dst.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(src, dst)
    return True


def check_go_environment() -> bool:
    """检查 Go 环境是否可用"""
    try:
        result = subprocess.run(
            ["go", "version"],
            capture_output=True,
            text=True,
            encoding="utf-8",
            errors="replace",
        )
        if result.returncode == 0:
            print(f"  {Console.info(t('go_version'))}: {result.stdout.strip()}")
            return True
    except FileNotFoundError:
        pass

    print(f"  {Console.err(t('error'))} {t('go_not_found')}")
    print()
    print(f"  {Console.info(t('go_install_prompt'))}")
    print(f"    - {Console.info(t('go_install_official'))}")
    print(f"    - {Console.info(t('go_install_windows'))}")
    print(f"    - {Console.info(t('go_install_macos'))}")
    print(f"    - {Console.info(t('go_install_linux'))}")
    print()
    print(f"  {Console.info(t('go_install_path'))}")
    return False


def build_go_agent(
    root_dir: Path,
    install_dir: Path,
    ci_mode: bool = False,
) -> bool:
    """构建 Go Agent"""
    if not check_go_environment():
        return False

    go_service_dir = root_dir / "agent" / "go-service"
    if not go_service_dir.exists():
        print(
            f"  {Console.err(t('error'))} {t('go_source_not_found')}: {go_service_dir}"
        )
        return False

    system = platform.system().lower()
    goos = {"windows": "windows", "darwin": "darwin"}.get(system, "linux")

    machine = platform.machine().lower()
    goarch = (
        "amd64"
        if machine in ("x86_64", "amd64")
        else "arm64"
        if machine in ("aarch64", "arm64")
        else machine
    )

    ext = ".exe" if goos == "windows" else ""

    agent_dir = install_dir / "agent"
    agent_dir.mkdir(parents=True, exist_ok=True)
    output_path = agent_dir / f"go-service{ext}"

    print(f"  {Console.info(t('target_platform'))}: {goos}/{goarch}")
    print(f"  {Console.info(t('output_path'))}: {output_path}")

    env = {**os.environ, "GOOS": goos, "GOARCH": goarch, "CGO_ENABLED": "0"}

    # go mod tidy
    # CI 模式下只校验是否已同步，避免静默改动依赖文件
    tidy_cmd = ["go", "mod", "tidy"]
    if ci_mode:
        tidy_cmd.append("-diff")

    tidy_result = subprocess.run(
        tidy_cmd,
        cwd=go_service_dir,
        capture_output=True,
        text=True,
        encoding="utf-8",
        env=env,
    )
    if tidy_result.stdout:
        print(tidy_result.stdout)
    if tidy_result.returncode != 0:
        if ci_mode:
            print(f"  {Console.err(t('error'))} {t('go_mod_files_out_of_sync')}")
            if tidy_result.stderr:
                max_stderr_chars = 8 * 1024
                stderr_snippet = tidy_result.stderr.rstrip()
                if len(stderr_snippet) > max_stderr_chars:
                    stderr_snippet = stderr_snippet[-max_stderr_chars:]
                print(
                    f"  {Console.err(t('error'))} {t('go_mod_tidy_stderr')}:\n{stderr_snippet}"
                )
        else:
            print(
                f"  {Console.err(t('error'))} {t('go_mod_tidy_failed')}: {tidy_result.stderr}"
            )
        return False
    if tidy_result.stderr:
        print(tidy_result.stderr)

    # go build
    # CI 模式：release with debug info（保留 DWARF 调试信息，不使用 -s -w）
    # 开发模式：debug 构建（保留调试信息 + 禁用优化，便于断点调试）
    if ci_mode:
        ldflags = ""
        gcflags = ""
    else:
        ldflags = ""
        gcflags = "all=-N -l"

    ldflags = ldflags.strip()

    build_cmd = ["go", "build"]

    if ci_mode:
        build_cmd.append("-trimpath")

    if gcflags:
        build_cmd.append(f"-gcflags={gcflags}")

    if ldflags:
        build_cmd.append(f"-ldflags={ldflags}")

    build_cmd.extend(["-o", str(output_path), "."])

    build_mode_text = t("build_mode_ci") if ci_mode else t("build_mode_dev")
    print(f"  {Console.warn(t('build_mode'))}: {build_mode_text}")
    print(f"  {Console.info(t('build_command'))}: {' '.join(build_cmd)}")

    result = subprocess.run(
        build_cmd,
        cwd=go_service_dir,
        capture_output=True,
        text=True,
        encoding="utf-8",
        env=env,
    )
    if result.stdout:
        print(result.stdout)
    if result.returncode != 0:
        print(f"  {Console.err(t('error'))} {t('go_build_failed')}:")
        if result.stderr:
            print(result.stderr)
        return False
    if result.stderr:
        print(result.stderr)

    print(f"  {Console.ok('->')} {output_path}")
    return True


def configure_ocr_model() -> bool:
    """Configure OCR model by copying from MaaCommonAssets.
    TODO: MaaEnd uses a git submodule for assets/resource/model instead of this.
    """
    assets_dir = PROJECT_BASE / "assets"
    assets_ocr_dir = assets_dir / "MaaCommonAssets" / "OCR"
    if not assets_ocr_dir.exists():
        print(Console.warn(t("wrn_ocr_assets_not_found", path=assets_ocr_dir)))
        return True  # Non-fatal

    ocr_dir = assets_dir / "resource" / "model" / "ocr"
    if not ocr_dir.exists():
        shutil.copytree(
            assets_dir / "MaaCommonAssets" / "OCR" / "ppocr_v5" / "zh_cn",
            ocr_dir,
            dirs_exist_ok=True,
        )
        print(Console.ok(t("inf_ocr_model_configured")))
    else:
        print(Console.ok(t("inf_ocr_model_exists")))
    return True


def main() -> None:
    init_local()

    parser = argparse.ArgumentParser(description=t("description"))
    parser.add_argument("--ci", action="store_true", help=t("arg_ci"))
    args = parser.parse_args()

    use_copy = args.ci

    assets_dir = PROJECT_BASE / "assets"
    install_dir = PROJECT_BASE / "install"

    mode_text = t("mode_ci") if use_copy else t("mode_dev")
    print(f"{Console.info(t('root_dir'))}: {PROJECT_BASE}")
    print(f"{Console.info(t('install_dir'))}:   {install_dir}")
    print(f"{Console.warn(t('mode'))}:       {mode_text}")
    print()

    install_dir.mkdir(parents=True, exist_ok=True)

    # 用于链接或复制的函数
    link_or_copy_dir = copy_directory if use_copy else create_directory_link
    link_or_copy_file = copy_file if use_copy else create_file_link

    # 1. 配置 OCR 模型
    #    TODO: MaaEnd uses a git submodule instead; remove this once model is a submodule.
    print(Console.step(t("step_configure_ocr")))
    if not configure_ocr_model():
        print(f"  {Console.err(t('error'))} {t('configure_ocr_failed')}")
        sys.exit(1)

    # 2. 链接/复制 assets 目录内容
    print(Console.step(t("step_process_assets")))
    for item in assets_dir.iterdir():
        # Skip MaaCommonAssets (submodule, not needed at runtime)
        #   and runtime-only directories (cache, debug) — they are
        #   regular dirs in install/ only.
        if item.name in ("MaaCommonAssets", "cache", "debug"):
            continue
        dst = install_dir / item.name
        if item.is_dir():
            if link_or_copy_dir(item, dst):
                print(f"  {Console.ok('->')} {dst}")
        elif item.is_file():
            if link_or_copy_file(item, dst):
                print(f"  {Console.ok('->')} {dst}")

    # 3. 构建 Go Agent
    print(Console.step(t("step_build_go")))
    if not build_go_agent(PROJECT_BASE, install_dir, ci_mode=use_copy):
        print(f"  {Console.err(t('error'))} {t('build_go_failed')}")
        sys.exit(1)

    # 4. 链接/复制项目根目录文件并创建 maafw 目录
    print(Console.step(t("step_prepare_files")))
    for filename in ["README.md", "LICENSE"]:
        src = PROJECT_BASE / filename
        dst = install_dir / filename
        if src.exists():
            if link_or_copy_file(src, dst):
                print(f"  {Console.ok('->')} {dst}")

    # 4a. Ensure runtime directories exist as regular dirs (not junctioned)
    for runtime_dir_name in ("cache", "debug"):
        runtime_dir = install_dir / runtime_dir_name
        runtime_dir.mkdir(parents=True, exist_ok=True)

    maafw_dir = install_dir / "maafw"
    maafw_dir.mkdir(parents=True, exist_ok=True)
    print(f"  {Console.ok('->')} {maafw_dir}")

    # Ensure MaaAgentBinary is present inside maafw
    agent_binary_src = PROJECT_BASE / "deps" / "share" / "MaaAgentBinary"
    if agent_binary_src.exists() and maafw_dir.exists():
        shutil.copytree(
            agent_binary_src,
            maafw_dir / "MaaAgentBinary",
            dirs_exist_ok=True,
        )

    print()
    print(t("separator"))
    print(Console.ok(t("install_complete")))

    if not use_copy:
        if not any(maafw_dir.iterdir()):
            print()
            print(Console.warn(t("maafw_download_hint")))
            print(f"  {t('maafw_download_step')}")
            print(f"  {t('maafw_download_url')}")
        if (
            not (install_dir / "mxu").exists()
            and not (install_dir / "mxu.exe").exists()
        ):
            print()
            print(Console.warn(t("mxu_download_hint")))
            print(f"  {t('mxu_download_step')}")
            print(f"  {t('mxu_download_url')}")

    print()


if __name__ == "__main__":
    main()
