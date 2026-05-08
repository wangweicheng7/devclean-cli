# devclean-cli

面向 macOS 软件开发者的终端清理工具，默认 **安全优先**（先预览，再清理）。

## 功能

- `scan` / `plan`：扫描可清理项并输出预览。
- `clean --dry-run`：仅预览，不会删除任何文件。
- `clean --confirm`：显式确认后执行删除。
- `clean --interactive`：逐项交互确认清理对象。
- `scan/clean --discover-projects`：自动发现开发工程目录，并纳入工程垃圾目录（如 `node_modules`、`dist`、`build`、`.dart_tool`、`ios/Pods`、`android/.gradle`）。
- `profile` 分层：`safe` / `dev` / `aggressive`（高风险项默认 `report_only`）。

### profile 怎么选？

- **safe**：只包含最保守、最不容易踩坑的项目（例如 Go build cache）。适合第一次试用。
- **dev（推荐日常默认）**：在 safe 基础上加入开发者常见项：工程垃圾（`node_modules`/`dist`/`build`/`.dart_tool`/`ios/Pods` 等），以及 **Xcode/Gradle 等默认 report-only 的“大项”**（显示但不删）。
- **aggressive**：保留给未来更激进/更可能影响环境的项目（当前一般不需要用它）。

### 清理 Xcode / Gradle / 用户缓存目录

- 默认 **report-only** 的项（如 Xcode DerivedData、Archives、`~/.gradle/caches`）不会随 `clean --confirm` 删除。
- 需要删除时，必须显式加上 **`--allow-report-only`**，并建议先 **`--dry-run`** 预览。
- **`--user-caches`**：把 `~/Library/Caches` 下**一级子目录**纳入扫描；同样默认 report-only，删除需配合 **`--allow-report-only`**。已单独列出的缓存（如 go-build、npm、Yarn、pip、CocoaPods）不会重复出现。

```bash
devclean scan --profile dev --user-caches --with-size
devclean clean --profile dev --user-caches --allow-report-only --dry-run
devclean clean --profile dev --user-caches --allow-report-only --confirm
```

## 快速开始（本地运行）

```bash
go run ./cmd/devclean scan --profile safe
go run ./cmd/devclean clean --profile dev --dry-run
go run ./cmd/devclean clean --profile dev --confirm
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

- `devclean scan [--profile safe|dev|aggressive] [--json] [--category cache,logs,build]`
- `devclean plan [同 scan]`
- `devclean clean [--dry-run] [--confirm] [--interactive] [--profile ...] [--category ...] [--json]`
- `devclean config init [--path .devcleanrc.json] [--force]`
- `devclean config prune-missing [--apply]`
- `devclean doctor`

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

