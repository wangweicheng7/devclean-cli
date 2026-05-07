package cli

import (
	"fmt"
	"io"

	"github.com/wangweicheng7/devclean-cli/internal/config"
)

func PrintHelp(w io.Writer) {
	fmt.Fprintf(w, `devclean - macOS developer cleanup CLI (safe-first)

Usage:
  devclean version
  devclean scan  [--config path] [--profile safe|dev|aggressive] [--category cache,logs,build] [--repo path] [--discover-projects] [--discover-roots a,b] [--discover-depth 4] [--discover-refresh] [--discover-debug] [--with-size] [--json]
  devclean plan  [same as scan]
  devclean clean [--config path] [--profile safe|dev|aggressive] [--category cache,logs,build] [--repo path] [--discover-projects] [--discover-roots a,b] [--discover-depth 4] [--discover-refresh] [--discover-debug] [--dry-run] [--confirm] [--interactive] [--with-size] [--json]
  devclean config init [--path path] [--force]
  devclean config exclude add|remove|list [--config path] [--dry-run] <id...>
  devclean config include add|remove|list [--config path] [--dry-run] <id...>
  devclean config prune-missing [--config path] [--apply]
  devclean doctor

Notes:
  - clean requires either --dry-run (preview) or --confirm (execute). If both are set, dry-run wins.
  - aggressive profile may include report-only items (not deleted).
  - config lookup: ./%s then ~/%s
`, config.DefaultConfigFilename, config.DefaultConfigFilename)
}
