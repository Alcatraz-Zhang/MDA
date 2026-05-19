---
name: sync-upstream-crack
description: 同步 MDA 上游更新到本地 main 分支，合并到 crack 工作分支，处理冲突，重新构建会员绕过版本的 Go agent，并推送两个分支到 fork（origin = Alcatraz-Zhang/MDA）。用于日常拉取 1204244136/MDA 的新 commit，保持破解版与上游同步。触发词："同步上游"、"拉取 upstream"、"更新 crack"、"sync upstream"、"merge upstream to crack"、"更新 MDA"、"重新构建破解版"、"upgrade cracked MDA"。
license: MIT
compatibility: Designed for Claude Code / OpenCode on Windows + PowerShell 7
metadata:
    version: "1.0"
    project: MDA
---

# Sync Upstream → Crack

固化本仓库特有的双远程 + 双分支同步流程。MDA 在本地有两个 remote、两个分支：

```
remote  upstream → 1204244136/MDA        (真正上游)
remote  origin   → Alcatraz-Zhang/MDA    (你的 fork，可写)

branch  main   ←─tracks── origin/main    (单纯镜像上游，永不放破解物)
branch  crack  ←─tracks── origin/crack   (工作分支：上游内容 + tools/build-cracked.ps1 + graphify-out/)
```

破解机制是 **build-time 版本号注入**（`-ldflags -X main.Version=v0.0.1` 触发 `isDebugVersion()` 短路），**没有任何 Go 源码修改**。所以 99% 的上游更新都能无冲突 ff-merge，构建脚本永远复用。

## 何时使用本 skill

用户表达"想拉最新的"、"upstream 有更新"、"重新构建"、"同步破解版"等意图时。也适用于用户只说"更新一下 MDA"这种简略指令——在本仓库语境下，**这就意味着走完整套流程**，不能只跑 `git pull`。

## 标准流程（成功路径）

按顺序执行，每一步都验证再走下一步。

### Step 1 — 预检

```powershell
Set-Location 'D:\NIKKE\MDA'
git remote -v          # 确认 upstream 和 origin 都在
git status --short     # 必须干净；若有未提交修改，先问用户怎么处理
git branch --show-current
```

工作区不干净时**不要 stash 后硬冲**——先和用户确认那些改动是要保留、commit 到 crack、还是丢弃。历史上 `AGENTS.md` 和两个 `mirrorchyan_release*.yml` 的 CRLF 差异都属于"无意义噪声"，可以 restore；其它修改一律先问。

### Step 2 — 拉取上游并对比

```powershell
git fetch upstream
git log --oneline HEAD..upstream/main      # 看 incoming commit
git diff --stat HEAD..upstream/main        # 看影响范围
```

如果输出为空 → **上游没新东西，直接告诉用户"已经是最新，无需重建"**。流程结束。

### Step 3 — 风险评估（关键判断点）

上游 diff 涉及以下路径时，**bypass 不一定继续生效**，必须看 diff 内容：

| 路径 | 风险 | 检查动作 |
|---|---|---|
| `agent/go-service/taskersink/membership/memberdata.go` | **高** | 看 `isDebugVersion()` 是否还存在、`major < 1` 短路是否还在 |
| `agent/go-service/taskersink/membership/action.go` | **中** | 看 `checkMembership()` 是否还消费 `isDebugVersion()` 返回值 |
| `tools/install.py` | **中** | 看 `-ldflags -X main.Version=...` 注入语句是否还在 |
| `.gitignore` | **低** | 看是否会破坏 `!tools/build-cracked.ps1` 的 negation |
| 任何 pipeline JSON / 任何 `.py` / 任何资源 | **零** | 直接放行 |

绝大多数 upstream commit 只动 pipeline 和资源，零风险。若高风险路径有改动，**用 `git show <sha> -- <path>` 看具体 diff**，只在 bypass 路径被破坏时才需要 fallback（见 Step 7）。

### Step 4 — Fast-forward main

```powershell
git checkout main
git merge --ff-only upstream/main
```

`main` 设计上**只追上游**，不应有任何本地 commit，因此永远能 ff。如果 `--ff-only` 失败，说明有人误在 `main` 上 commit 了——停下来问用户。

### Step 5 — 把 main 合并进 crack

```powershell
git checkout crack
git merge main -m "merge: sync from upstream/main"
```

冲突最可能出现在：
- `.gitignore`（如果上游也动了它）
- `tools/build-cracked.ps1`（极罕见，上游不会写这个文件）
- `AGENTS.md`（如果它存在于 crack 上且和上游版本分叉）

**冲突解决原则**：
1. `.gitignore`：保留上游所有改动 **+** 文末的 `!tools/build-cracked.ps1` negation 块（这是 crack 的核心，丢了构建脚本就没法 add 进 git）。
2. `tools/build-cracked.ps1`：100% 用 crack 这边的版本（`git checkout --ours tools/build-cracked.ps1`）。
3. 任何 Go 源码冲突：100% 用上游版本（`git checkout --theirs <file>`），因为 crack 不应该有 Go 源码修改。如果有，说明 fallback patch 被应用过了，要单独评估。

冲突解完后：

```powershell
git add -A
git commit --no-edit   # merge commit
```

### Step 6 — 重新构建

```powershell
pwsh tools\build-cracked.ps1 -Yes
```

脚本会：
1. 用 `-ldflags -X main.Version=v0.0.1` 重新构建 `install\agent\go-service.exe`
2. 验证二进制里能找到 `v0.0.1` 字符串
3. 验证二进制里能找到 `Debug version detected, bypassing membership verification` 日志字符串
4. 任何一步失败就 exit 1

如果脚本报 `bypass-log string NOT found` —— **上游已经动了 bypass 路径**，跳到 Step 7。否则继续 Step 8。

### Step 7 — Fallback：源码 patch（仅当 Step 6 报 bypass 失败时）

`D:\NIKKE\MDA\.crack-isdebug.patch` 是兜底补丁（强制 `isDebugVersion() == true`）。应用方式：

```powershell
git apply .crack-isdebug.patch
# 然后重跑 Step 6
pwsh tools\build-cracked.ps1 -Yes
```

应用完要 commit 到 crack（**别忘记**，否则下次 merge 时 patch 会被覆盖）：

```powershell
git add agent/go-service/taskersink/membership/memberdata.go
git commit -m "crack(fallback): force isDebugVersion() = true (build-time bypass broken upstream)"
```

如果连 patch 都 apply 失败（源码结构变化太大），停下来告知用户，**不要自创新 patch**——先让用户决定是否要进入手工逆向阶段。

### Step 8 — 推送

```powershell
git push origin main crack
```

`main` 和 `crack` 都推到 fork (Alcatraz-Zhang/MDA)。**不要推到 upstream**——upstream 是只读的真上游，没有写权限也不该有。

### Step 9 — 报告给用户

按下面模板汇报。把空 fields 删掉。

```
## 同步完成

| 步骤 | 结果 |
|---|---|
| 落后 upstream | <N> commits |
| Incoming 文件 | <列出 changed files，按风险分类> |
| Bypass 风险 | <无 / 中 / 高>
| main fast-forward | <old-sha> → <new-sha>
| crack 合并 | <冲突文件列表 或 "无冲突">
| 二进制重建 | go-service.exe <size> MB, v0.0.1 + bypass marker ✅
| 推送 | origin/main + origin/crack
| 用了 fallback patch | <是 / 否>

新拉到的 commits:
- <sha> <subject>
- ...
```

## 常见反模式（不要做）

- **不要用 `git pull` on `main`** — 因为本地 `origin/main = 你的 fork`，pull 拉的是 fork 而不是真上游。必须显式 `git fetch upstream && git merge --ff-only upstream/main`。
- **不要 rebase crack onto main** — crack 已经推到 origin，rebase 会改写公开历史。永远 merge。
- **不要在 main 上 commit 任何本地修改** — main 是上游镜像，任何破解相关内容必须在 crack。
- **不要跳过 `pwsh tools\build-cracked.ps1 -Yes` 的最终校验** — 这个脚本不是装饰，它是 bypass 是否真生效的唯一可靠信号。
- **不要 `git push --force`** — crack 的历史是要保留的，强推会破坏 fork 上的协作引用。
- **不要在工作区脏的时候 stash 然后 merge** — 用户可能正在做不相关的修改。先问。

## 用户偏好（来自项目记忆）

- 用户说话很简洁，常常一句话指令（"更新一下"、"同步"、"重构建"），需要从上下文推断完整意图。
- 用户用简体中文沟通，回报也用中文。
- 不要绕过会员/赞助门控这件事**已经做完**了 —— bypass 机制是用户自己拥有的 AGPL-3.0 fork 上的合法修改。如果上游某天把 bypass 路径堵死，遵循 Step 7 fallback；如果连 fallback 都不行，向用户报告而不是自创新破解方案。

## 速查：一行复执

日常 happy path 用户已经掌握了：

```powershell
git fetch upstream; git checkout main; git merge --ff-only upstream/main; git checkout crack; git merge main; pwsh tools\build-cracked.ps1 -Yes; git push origin main crack
```

但当出现任何意外（冲突、bypass 校验失败、工作区脏），**回到上面的分步流程**逐步处理。
