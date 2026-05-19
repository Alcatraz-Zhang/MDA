#!/usr/bin/env pwsh
<#
.SYNOPSIS
  One-click cracked build of MDA via version-string trick (no source mod).

.DESCRIPTION
  Builds the Go agent with main.Version=v0.x.x via -ldflags. That makes
  isDebugVersion() in agent/go-service/taskersink/membership/memberdata.go
  return true, which short-circuits checkMembership() to always return
  IsMember=true without contacting doropay.top — unlocking all 🍊
  sponsor-only tasks (Arena, Tribe Tower, Advise, Coordinated Operations,
  Large/Small Event).

  No source code modifications required. Survives upstream merges as long
  as the author keeps the existing isDebugVersion() / checkMembership()
  debug short-circuit.

.PARAMETER Version
  Version string injected via -ldflags. Triggers debug-bypass when:
    - v0.x.x  (major < 1)
    - dev / empty / non-numeric  (parse failure)
  Default: v0.0.1.

.PARAMETER OS
  Target OS: win | macos | linux. Default: win.

.PARAMETER Arch
  Target arch: x86_64 | aarch64. Default: x86_64.

.PARAMETER Full
  Run full install (copy resources + interface + agent). Default: build
  Go agent only (--build-go-only).

.PARAMETER Yes
  Non-interactive: use defaults for all unspecified parameters.

.PARAMETER Force
  Allow building a non-bypass binary (Version >= v1.0.0). Without this
  switch, the script refuses to build a binary that will hit the
  membership gate at runtime.

.EXAMPLE
  .\tools\build-cracked.ps1
  Interactive build, all defaults.

.EXAMPLE
  .\tools\build-cracked.ps1 -Yes
  Fast non-interactive Go-only rebuild with v0.0.1.

.EXAMPLE
  .\tools\build-cracked.ps1 -Full -Yes
  Full install (resources + agent), non-interactive.

.EXAMPLE
  .\tools\build-cracked.ps1 -Version v1.2.3 -Force -Yes
  Build a NON-bypass binary (membership gate active). Rare; mostly for
  reproducing official release behavior.
#>
[CmdletBinding()]
param(
    [string]$Version = '',
    [ValidateSet('win', 'macos', 'linux')][string]$OS = 'win',
    [ValidateSet('x86_64', 'aarch64')][string]$Arch = 'x86_64',
    [switch]$Full,
    [switch]$Yes,
    [switch]$Force
)

$ErrorActionPreference = 'Stop'
$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path

# Force UTF-8 console output so the zh-CN summary banner renders correctly
# on the default Windows code page (cp936/cp437). Mirrors scripts\utf8-cmd.cmd.
try {
    $OutputEncoding = [Text.UTF8Encoding]::new()
    [Console]::OutputEncoding = [Text.UTF8Encoding]::new()
} catch {}

function Write-Step ($m) { Write-Host "==> $m" -ForegroundColor Cyan }
function Write-Ok   ($m) { Write-Host "    [OK] $m" -ForegroundColor Green }
function Write-Warn2($m) { Write-Host "    [! ] $m" -ForegroundColor Yellow }
function Write-Err2 ($m) { Write-Host "    [XX] $m" -ForegroundColor Red }

Write-Host ""
Write-Host "===========================================" -ForegroundColor Magenta
Write-Host " MDA Cracked Build (zero source mod)"        -ForegroundColor Magenta
Write-Host " Trick: ldflags v0.x.x => isDebugVersion=true" -ForegroundColor DarkGray
Write-Host "===========================================" -ForegroundColor Magenta
Write-Host ""

# === Pre-flight: required toolchain ===
Write-Step "Pre-flight checks"
$missing = @()
foreach ($t in 'python', 'go') {
    $c = Get-Command $t -ErrorAction SilentlyContinue
    if (-not $c) {
        $missing += $t
        Write-Err2 "$t not on PATH"
    } else {
        Write-Ok "$t : $($c.Source)"
    }
}
if ($missing.Count -gt 0) {
    Write-Err2 "Install missing tools and retry."
    exit 1
}

# === Interactive prompts (skipped with -Yes) ===
if (-not $Yes) {
    if ([string]::IsNullOrWhiteSpace($Version)) {
        $in = Read-Host "Version (default v0.0.1, MUST be v0.x.x for bypass)"
        $Version = if ([string]::IsNullOrWhiteSpace($in)) { 'v0.0.1' } else { $in.Trim() }
    }
    if (-not $PSBoundParameters.ContainsKey('Full')) {
        $in = Read-Host "Mode: [g]o-only fast | [f]ull install (default g)"
        if ($in -match '^[Ff]') { $Full = $true }
    }
}
if ([string]::IsNullOrWhiteSpace($Version)) { $Version = 'v0.0.1' }

# === Validate the version actually triggers the debug bypass ===
# isDebugVersion() logic in memberdata.go:
#   - parse appVersion as semver "vMAJOR.MINOR.PATCH"
#   - parse failure (empty, "dev", "fork-2024", etc.) => return true
#   - parse success && major < 1 => return true
#   - parse success && major >= 1 => return false
$debugTrigger = $false
if ($Version -match '^v?(\d+)\.(\d+)\.(\d+)') {
    if ([int]$Matches[1] -lt 1) { $debugTrigger = $true }
} elseif ($Version -match '^v?(\d+)\.(\d+)$') {
    if ([int]$Matches[1] -lt 1) { $debugTrigger = $true }
} elseif ($Version -match '^v?(\d+)$') {
    if ([int]$Matches[1] -lt 1) { $debugTrigger = $true }
} else {
    # non-numeric -> isDebugVersion strconv.Atoi parse fails -> returns true
    $debugTrigger = $true
}

if (-not $debugTrigger) {
    Write-Warn2 "Version '$Version' has major>=1; membership check WILL run."
    Write-Warn2 "Cracked bypass NOT active for this build."
    if (-not $Force) {
        Write-Err2 "Refusing to build a non-bypass binary."
        Write-Err2 "Use -Force to override, or pass -Version v0.0.1 (default) for the bypass build."
        exit 1
    }
    Write-Warn2 "-Force given; continuing anyway."
}

# === Show the build plan ===
Write-Host ""
Write-Step "Build plan"
$mode = if ($Full) { 'full install (resources + interface + agent)' } else { 'go-only (--build-go-only)' }
$bp = if ($debugTrigger) { 'YES (isDebugVersion=true)' } else { 'NO (membership check active)' }
Write-Host "  version : $Version"
Write-Host "  os/arch : $OS / $Arch"
Write-Host "  mode    : $mode"
Write-Host "  bypass  : $bp"
Write-Host ""

# === deps/bin check for full install ===
if ($Full) {
    $depsBin = Join-Path $RepoRoot 'deps\bin'
    if (-not (Test-Path $depsBin)) {
        Write-Err2 "deps\bin missing. Run first: python tools\setup_workspace.py"
        exit 1
    }
    Write-Ok "deps\bin present"

    # tools\setup_workspace.py creates junctions and hardlinks in install\ that
    # point back into the repo (install\maafw -> deps\bin, install\resource ->
    # assets\resource, install\interface.json hardlinked to assets\interface.json,
    # etc). install.py -Full uses shutil.copytree/copy2 which tries to copy these
    # files onto themselves through the link, producing WinError 32 (file in use).
    # Detect and remove the dev-mode links so install.py can write real copies.
    # Only repo-internal links are removed — user data is untouched.
    $installPath = Join-Path $RepoRoot 'install'
    if (Test-Path $installPath) {
        $links = Get-ChildItem -Path $installPath -Force -ErrorAction SilentlyContinue |
            Where-Object { $_.LinkType -in 'Junction', 'HardLink', 'SymbolicLink' }
        if ($links) {
            Write-Step "Cleaning $($links.Count) dev-mode link(s) from setup_workspace.py"
            foreach ($l in $links) {
                $tgt = if ($l.Target) { $l.Target -join ';' } else { '(repo internal)' }
                Write-Host "    remove $($l.LinkType): $($l.Name) -> $tgt"
                if ($l.LinkType -eq 'Junction') {
                    # rmdir unlinks the junction without touching the target
                    cmd /c rmdir "`"$($l.FullName)`"" 2>&1 | Out-Null
                } else {
                    # HardLink/SymbolicLink: Remove-Item unlinks just this entry;
                    # the file content remains accessible through other names.
                    Remove-Item -LiteralPath $l.FullName -Force -ErrorAction SilentlyContinue
                }
                if (Test-Path -LiteralPath $l.FullName) {
                    Write-Err2 "Failed to remove $($l.LinkType) $($l.FullName)"
                    exit 1
                }
            }
            Write-Ok "links cleared"
        }
    }
}

# === Invoke install.py ===
Write-Step "Invoking tools\install.py"
$pyArgs = @()
if (-not $Full) { $pyArgs += '--build-go-only' }
$pyArgs += $Version, $OS, $Arch
Write-Host "    cmd: python tools\install.py $($pyArgs -join ' ')" -ForegroundColor DarkGray

Push-Location $RepoRoot
try {
    & python tools\install.py @pyArgs
    if ($LASTEXITCODE -ne 0) { throw "install.py failed (exit $LASTEXITCODE)" }
} finally {
    Pop-Location
}
Write-Ok "install.py exit 0"

# === Verify the binary was produced ===
Write-Step "Verifying output binary"
$ext = if ($OS -eq 'win') { '.exe' } else { '' }
$exe = Join-Path $RepoRoot "install\agent\go-service$ext"
if (-not (Test-Path $exe)) {
    Write-Err2 "binary not found: $exe"
    exit 1
}

$f = Get-Item $exe
Write-Ok "path  : $($f.FullName)"
Write-Ok "size  : $([Math]::Round($f.Length / 1MB, 2)) MB"
Write-Ok "built : $($f.LastWriteTime)"

# === Sanity-check: scan the binary for our key strings ===
$bytes = [IO.File]::ReadAllBytes($exe)
$asStr = [Text.Encoding]::ASCII.GetString($bytes)

$marker = 'Debug version detected, bypassing membership verification'
if ($asStr.Contains($marker)) {
    Write-Ok "bypass-log string present in binary"
} else {
    Write-Warn2 "bypass-log string NOT found (compiler may have inlined; bypass should still work if logic intact)"
}

if ($asStr.Contains($Version)) {
    Write-Ok "version '$Version' embedded in binary"
} else {
    Write-Warn2 "version '$Version' not literally present (may be embedded differently)"
}

# === Summary banner ===
Write-Host ""
Write-Host "===========================================" -ForegroundColor Green
Write-Host " BUILD COMPLETE   bypass: $bp"               -ForegroundColor Green
Write-Host "===========================================" -ForegroundColor Green
Write-Host "  Run:  cd install ; .\MaaPiCli.exe"         -ForegroundColor DarkGray
Write-Host "  Tasks: 竞技场 / 无限之塔 / 咨询 / 协同作战 / 大型活动 / 小型活动 should now work" -ForegroundColor DarkGray
Write-Host ""
