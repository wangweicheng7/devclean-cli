package cli

import (
	"fmt"
	"io"
)

func PrintHelp(w io.Writer) {
	fmt.Fprint(w, `cleandev - macOS developer cleanup CLI (safe-first)

Usage:
  cleandev scan  [--profile safe|dev|aggressive] [--category cache,logs,build] [--repo path] [--with-size] [--json]
  cleandev plan  [same as scan]
  cleandev clean [--profile safe|dev|aggressive] [--category cache,logs,build] [--repo path] [--dry-run] [--confirm] [--with-size] [--json]
  cleandev doctor

Notes:
  - clean requires either --dry-run (preview) or --confirm (execute). If both are set, dry-run wins.
  - aggressive profile may include report-only items (not deleted).
`)
}

