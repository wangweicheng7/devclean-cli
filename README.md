# devclean-cli

面向 macOS 软件开发者的终端清理工具，默认 **安全优先**（先预览，再清理）。

## 功能

- `scan` / `plan`：扫描可清理项并输出预览。
- `clean --dry-run`：仅预览，不会删除任何文件。
- **`devclean clean`（默认）**：先**扫描**并打印与 `scan` 类似的汇总表，再提示一次 **`proceed with deletion? [y/N]`**，输入 `y` 后才真正删除（无需其它确认旗标）。
- `clean --interactive`：按**工程分组**逐项交互确认；同一工程下多项（如 `node_modules`、`dist`、`build`）一次 `y` 可一并处理；输入 `y` 后立即生效。
- `scan` / `plan` / `clean` 支持 **`--all`**：默认会隐藏「空目录」（`bytes=0` 且无文件），加 `--all` 则全部列出。
- `scan/clean --discover-projects`：自动发现开发工程目录，并纳入工程垃圾目录（如 `node_modules`、`dist`、`build`、`.dart_tool`、`ios/Pods`、`android/.gradle`）。
- **`--user-caches`**：扫描 `~/Library/Caches` 下**一级子目录**（已单独列出的 go-build、npm、Yarn、pip、CocoaPods 等不会重复出现）；**不需要写 `--profile`**，条目在默认 `safe` profile 下即可参与扫描。
- `profile` 分层：`safe` / `dev` / `aggressive`（部分高风险项仍为 `report_only`，见下文）。
- `devclean version`：查看版本号（亦支持 `devclean --version` / `-v`）。

### profile 怎么选？

- **safe**：只包含最保守、最不容易踩坑的项目（例如 Go build cache）。适合第一次试用。
- **dev（推荐日常默认）**：在 safe 基础上加入开发者常见项：工程垃圾（`node_modules`/`dist`/`build`/`.dart_tool`/`ios/Pods` 等），以及 **Xcode DerivedData（默认可删）**；**Xcode Archives、`~/.gradle/caches` 等仍为 report-only**，需 `--allow-report-only` 才允许删除。
- **aggressive**：保留给未来更激进/更可能影响环境的项目（当前一般不需要用它）。

### 用户缓存 `~/Library/Caches`

- **`devclean scan --user-caches`**：列出用户缓存一级子目录体积，**不必加 `--profile`**。
- **`devclean clean --user-caches`**：与其它 `clean` 一样，先扫描再 **`y/N`** 总确认；建议先 **`--dry-run`** 预览。
- 若仍希望用 **`--allow-report-only`** 去删其它 report-only 大项（如 Archives、Gradle 用户缓存目录），可与 `--user-caches` 组合使用。

```bash
devclean scan --user-caches --with-size
devclean clean --user-caches --dry-run
devclean clean --user-caches
```

### 清理其它 report-only 大项（Xcode Archives / Gradle 等）

- 默认 **report-only** 的项不会在未加 **`--allow-report-only`** 时被删除。
- 需要删除时，显式加 **`--allow-report-only`**，并建议先 **`--dry-run`**。

```bash
devclean scan --profile dev --with-size
devclean clean --profile dev --allow-report-only --dry-run
devclean clean --profile dev --allow-report-only
```

## 快速开始（本地运行）

```bash
go run ./cmd/devclean scan --profile safe
go run ./cmd/devclean clean --profile dev --dry-run
go run ./cmd/devclean clean --profile dev
```

## 安装（Homebrew）

本仓库包含 tap 公式模板：`homebrew-tap/Formula/devclean-cli.rb`。

你发布 release 后，将公式里的：
- `url`
- `sha256`
- `version`

替换为对应 release 的值，然后将 formula 推到独立 tap 仓库（例如 `wangweicheng7/homebrew-tap`）。

可使用脚本自动更新 formula（从 release 的 `checksums.txt` 读取 arm64/amd64 校验）：

```bash
make brew-formula-update TAG=v0.2.0
```

并可一键提交并推送到 tap 仓库：

```bash
make brew-formula-publish
```

推荐安装方式（已 tap 后使用简写）：

```bash
brew tap wangweicheng7/homebrew-tap
brew install devclean-cli
devclean doctor
```

如果需要排障，可使用全限定写法：

```bash
brew install wangweicheng7/homebrew-tap/devclean-cli
```

## 可选：添加短别名 `dcl`

我们不在安装阶段自动修改你的 shell 配置文件（如 `~/.zshrc` / `~/.bashrc`），避免污染用户环境。你可以手动添加：

zsh:

```bash
echo "alias dcl='devclean'" >> ~/.zshrc
source ~/.zshrc
```

bash:

```bash
echo "alias dcl='devclean'" >> ~/.bashrc
source ~/.bashrc
```

## 命令

与 `devclean --help` 一致，常用形式如下：

- `devclean version`（或 `devclean -v` / `--version`）
- `devclean scan [--config path] [--profile safe|dev|aggressive] [--category cache,logs,build] [--repo path] [--discover-projects] [--discover-roots a,b] [--discover-depth N] [--discover-refresh] [--discover-debug] [--user-caches] [--all] [--with-size] [--json]`
- `devclean plan`：参数同 `scan`
- `devclean clean`：在 `scan` 基础上还可加 `[--allow-report-only] [--dry-run] [--interactive] [--interactive-batch] [--with-size] [--json]`
- `devclean config init [--path path] [--force]`
- `devclean config exclude add|remove|list [--config path] [--dry-run] <id...>`
- `devclean config include add|remove|list [--config path] [--dry-run] <id...>`
- `devclean config prune-missing [--config path] [--apply]`
- `devclean doctor`

说明：非交互的 **`clean`** 在真正删除前会**先扫描并展示计划**，再 **`y/N`** 一次；**`--dry-run`** 只预览、不提示删除；**`--interactive`** 用逐项选择代替这一次总确认。

示例（扫描常见代码目录中的工程垃圾）：

```bash
devclean scan --profile dev --discover-projects --discover-roots ~/Code,~/Projects --with-size
devclean clean --discover-projects --discover-roots ~/Code,~/Projects --dry-run
# 强制刷新发现缓存
devclean scan --discover-projects --discover-refresh
# 查看发现调试日志
devclean scan --discover-projects --discover-debug
```

## 配置文件（可选）

默认配置文件名：`.devcleanrc.json`（当前目录）。

查找顺序：**当前目录 `./.devcleanrc.json` > 用户目录 `~/.devcleanrc.json`**。

优先级：**CLI 参数 > `--config` 指定文件 > 自动发现的配置文件 > 默认值**。

建议先生成一份配置，把日常默认的 profile 固化下来：

```bash
devclean config init
```

可按 target `id` 排除或强制包含条目：`devclean config exclude …`、`devclean config include …`（见 `devclean --help`）。

