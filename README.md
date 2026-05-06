# devclean-cli

面向 macOS 软件开发者的终端清理工具，默认 **安全优先**（先预览，再清理）。

## 功能

- `scan` / `plan`：扫描可清理项并输出预览。
- `clean --dry-run`：仅预览，不会删除任何文件。
- `clean --confirm`：显式确认后执行删除。
- `clean --interactive`：逐项交互确认清理对象。
- `profile` 分层：`safe` / `dev` / `aggressive`（高风险项默认 `report_only`）。

## 快速开始（本地运行）

```bash
go run ./cmd/cleandev scan --profile safe
go run ./cmd/cleandev clean --profile dev --dry-run
go run ./cmd/cleandev clean --profile dev --confirm
```

## 安装（Homebrew）

本仓库包含 tap 公式模板：`homebrew-tap/Formula/devclean-cli.rb`。

你发布 release 后，将公式里的：
- `url`
- `sha256`
- `version`

替换为对应 release 的值，然后将 formula 推到独立 tap 仓库（例如 `wangweicheng7/homebrew-tap`）。

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
- `devclean doctor`

