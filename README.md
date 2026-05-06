# cleandev-cli

面向 macOS 软件开发者的终端清理工具，默认 **安全优先**（先预览，再清理）。

## 功能

- `scan` / `plan`：扫描可清理项并输出预览。
- `clean --dry-run`：仅预览，不会删除任何文件。
- `clean --confirm`：显式确认后执行删除。
- `profile` 分层：`safe` / `dev` / `aggressive`（高风险项默认 `report_only`）。

## 快速开始（本地运行）

```bash
go run ./cmd/cleandev scan --profile safe
go run ./cmd/cleandev clean --profile dev --dry-run
go run ./cmd/cleandev clean --profile dev --confirm
```

## 安装（Homebrew）

本仓库包含 tap 公式模板：`homebrew-tap/Formula/cleandev.rb`。

你发布 release 后，将公式里的：
- `url`
- `sha256`
- `version`

替换为对应 release 的值，然后将 formula 推到独立 tap 仓库（例如 `wangweicheng7/homebrew-tap`），用户即可安装：

```bash
brew tap wangweicheng7/tap
brew install cleandev
cleandev doctor
```

## 命令

- `cleandev scan [--profile safe|dev|aggressive] [--json] [--category cache,logs,build]`
- `cleandev plan [同 scan]`
- `cleandev clean [--dry-run] [--confirm] [--profile ...] [--category ...] [--json]`
- `cleandev doctor`

