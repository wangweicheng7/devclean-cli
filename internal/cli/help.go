package cli

import (
	"fmt"
	"io"
)

func PrintHelp(w io.Writer) {
	fmt.Fprint(w, `devclean - macOS developer cleanup CLI (safe-first)

Usage:
  devclean scan  [--config path] [--profile safe|dev|aggressive] [--category cache,logs,build] [--repo path] [--with-size] [--json]
  devclean plan  [same as scan]
  devclean clean [--config path] [--profile safe|dev|aggressive] [--category cache,logs,build] [--repo path] [--dry-run] [--confirm] [--interactive] [--with-size] [--json]
  devclean config init [--path .devcleanrc.json] [--force]
  devclean doctor

Notes:
  - clean requires either --dry-run (preview) or --confirm (execute). If both are set, dry-run wins.
  - aggressive profile may include report-only items (not deleted).
`)
}

