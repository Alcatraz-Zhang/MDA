# Project Agent Instructions

## Platform & Environment

- **Operating System**: Windows (this project is Windows-only)
- **Shell**: PowerShell 7
- **Path Separator**: Use backslash `\` for paths (e.g., `C:\Users\...`), not forward slash `/`

## Terminal Environment

- This project runs terminal commands in **PowerShell 7** on Windows.
- When executing commands, write PowerShell 7-compatible syntax and quote paths that contain spaces.
- Avoid Bash-only syntax such as heredocs (`cat <<'EOF'`), Unix pipelines that rely on GNU tools, and commands like `grep`, `find`, `sed`, or `awk` unless explicitly required and available.
- Avoid Linux/macOS-specific commands (e.g., `ls`, `cat`, `chmod`, `mkdir -p`, `rm -rf`).
- Prefer PowerShell-native commands, modern PowerShell 7 features, or repository-provided npm/scripts commands to avoid shell-syntax trial and error.
- Do not switch to `cmd`, Git Bash, WSL Bash, or other shells unless the task explicitly requires it.

## Canonical Crack Sync Workflow

- The single source of truth for upstream-to-crack synchronization is `.agents\skills\MDA-sync-crack\SKILL.md`.
- Future crack-sync automation must load and follow that `.agents` skill; do not copy its workflow into `.claude` or another agent-specific directory.
- `.claude\skills\MDA-sync-crack\SKILL.md` is only a compatibility redirect. If it conflicts with `.agents\skills\MDA-sync-crack\SKILL.md`, the `.agents` file wins.
- The final build and verification entrypoint remains `pwsh tools\build-cracked.ps1`; `tools\install.py` is only an internal dependency or debugging aid.

## Common PowerShell Equivalents

| Linux/macOS | PowerShell                                      |
| ----------- | ----------------------------------------------- |
| `ls`        | `Get-ChildItem` or `dir`                        |
| `cat`       | `Get-Content` or `type`                         |
| `mkdir -p`  | `New-Item -ItemType Directory -Force`           |
| `rm -rf`    | `Remove-Item -Recurse -Force`                   |
| `cp -r`     | `Copy-Item -Recurse`                            |
| `mv`        | `Move-Item`                                     |
| `chmod`     | `Set-ItemProperty` or `icacls`                  |
| `grep`      | `Select-String`                                 |
| `find`      | `Get-ChildItem -Recurse`                        |
| `sed`       | `-replace` operator or `ForEach-Object`         |
| `awk`       | `ConvertFrom-Csv`, `Select-Object`, or `-split` |
