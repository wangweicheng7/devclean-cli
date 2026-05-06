package cli

import (
	"os"
	"path/filepath"
	"strings"
)

func pruneMissingProjectIDs(ids []string) (removed []string, kept []string) {
	for _, id := range ids {
		root, ok := projectRootFromID(id)
		if !ok {
			kept = append(kept, id)
			continue
		}
		// Only prune project IDs that embed a path.
		if root == "" || !filepath.IsAbs(root) {
			kept = append(kept, id)
			continue
		}
		if strings.Contains(root, string(filepath.Separator)+".git"+string(filepath.Separator)) || strings.HasSuffix(root, string(filepath.Separator)+".git") {
			kept = append(kept, id)
			continue
		}
		if st, err := os.Stat(root); err == nil && st.IsDir() {
			kept = append(kept, id)
			continue
		}
		removed = append(removed, id)
	}
	return removed, kept
}

