package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wangweicheng7/devclean-cli/internal/clean"
)

func RunScan(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	fs.SetOutput(errOut)

	configPath := &stringFlag{}
	fs.Var(configPath, "config", "path to config file (default: .devcleanrc.json in current dir; fallback: ~/.devcleanrc.json)")

	profile := &stringFlag{v: string(clean.ProfileSafe)}
	fs.Var(profile, "profile", "safe|dev|aggressive")

	category := &stringFlag{}
	fs.Var(category, "category", "comma-separated: cache,logs,build (default all)")

	repo := &stringFlag{}
	fs.Var(repo, "repo", "path to repository root (scan build artifacts; report-only)")
	discoverProjects := newBoolFlag(false)
	fs.Var(discoverProjects, "discover-projects", "discover dev project folders under roots and include project junk dirs")
	discoverRoots := &stringFlag{}
	fs.Var(discoverRoots, "discover-roots", "comma-separated roots for project discovery (default ~/Code,~/Projects,~/workspace)")
	discoverDepth := &intFlag{v: 4}
	fs.Var(discoverDepth, "discover-depth", "max directory depth for project discovery")
	discoverRefresh := fs.Bool("discover-refresh", false, "force refresh project discovery cache")
	discoverDebug := fs.Bool("discover-debug", false, "print project discovery debug logs")
	userCaches := fs.Bool("user-caches", false, "include ~/Library/Caches/* (top-level only, report-only by default)")
	all := fs.Bool("all", false, "include all candidates, including empty directories")

	asJSON := fs.Bool("json", false, "output as json")

	withSizeFlag := newBoolFlag(true)
	fs.Var(withSizeFlag, "with-size", "calculate directory sizes (may be slow)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	cfg, usedPath, err := loadConfig(configPath.v)
	if err != nil {
		fmt.Fprintln(errOut, validateConfigPath(err, usedPath).Error())
		return 1
	}
	applyStringFromConfig(&profile.v, profile.IsSet(), cfg.Profile)
	applyStringFromConfig(&category.v, category.IsSet(), cfg.Category)
	applyStringFromConfig(&repo.v, repo.IsSet(), cfg.Repo)
	applyBoolFromConfig(&withSizeFlag.v, withSizeFlag.IsSet(), cfg.WithSize)
	if cfg.Discover != nil {
		applyBoolFromConfig(&discoverProjects.v, discoverProjects.IsSet(), cfg.Discover.Enabled)
		if !discoverRoots.IsSet() && len(cfg.Discover.Roots) > 0 {
			discoverRoots.v = strings.Join(cfg.Discover.Roots, ",")
		}
		applyIntFromConfig(&discoverDepth.v, discoverDepth.IsSet(), cfg.Discover.MaxDepth)
	}

	p, err := clean.ParseProfile(profile.v)
	if err != nil {
		fmt.Fprintln(errOut, err.Error())
		return 2
	}
	if discoverProjects.v && p == clean.ProfileSafe {
		fmt.Fprintln(errOut, "tip: --discover-projects only includes project junk targets on --profile dev|aggressive (current: safe)")
	}

	cats, err := clean.ParseCategories(category.v)
	if err != nil {
		fmt.Fprintln(errOut, err.Error())
		return 2
	}
	catSet := map[clean.Category]bool{}
	for _, c := range cats {
		catSet[c] = true
	}
	if len(catSet) == 0 {
		catSet = nil
	}

	var discoverLogs []string
	plan, err := clean.BuildPlan(ctx, clean.ScanOptions{
		Profile:    p,
		Categories: catSet,
		WithSize:   withSizeFlag.v,
		All:        *all,
		RepoRoot:   repo.v,
		UserCaches: *userCaches,
		Discover: clean.DiscoverOptions{
			Enabled:   discoverProjects.v,
			Roots:     splitCSV(discoverRoots.v),
			MaxDepth:  discoverDepth.v,
			Refresh:   *discoverRefresh,
			Debug:     *discoverDebug,
			DebugLogs: &discoverLogs,
		},
	})
	if err != nil {
		fmt.Fprintln(errOut, err.Error())
		return 1
	}

	if len(cfg.IncludeIDs) > 0 || len(cfg.ExcludeIDs) > 0 {
		plan.Items = filterItemsByIDs(plan.Items, cfg.IncludeIDs, cfg.ExcludeIDs)
	}

	if *asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		_ = enc.Encode(plan)
		return 0
	}
	if *discoverDebug {
		for _, l := range discoverLogs {
			fmt.Fprintf(errOut, "[discover] %s\n", l)
		}
	}

	printPlanTable(out, plan)
	return 0
}

func filterItemsByIDs(items []clean.Item, includeIDs, excludeIDs []string) []clean.Item {
	includeSet := map[string]bool{}
	excludeSet := map[string]bool{}
	for _, id := range includeIDs {
		includeSet[id] = true
	}
	for _, id := range excludeIDs {
		excludeSet[id] = true
	}
	var out []clean.Item
	for _, it := range items {
		if excludeSet[it.ID] {
			continue
		}
		if len(includeSet) > 0 && !includeSet[it.ID] {
			continue
		}
		out = append(out, it)
	}
	return out
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func printPlanTable(w io.Writer, plan clean.Plan) {
	home, _ := os.UserHomeDir()

	type projRow struct {
		name string
		path string
		b    int64
	}

	// Collect items that actually exist on disk. This keeps output focused.
	type row struct {
		name string
		b    int64
	}
	var rows []row
	projMap := map[string]*projRow{}
	var total int64
	maxName := 0

	for _, it := range plan.Items {
		if !it.Exists || it.Skipped {
			continue
		}
		if root, ok := projectRootFromID(it.ID); ok {
			pr, exists := projMap[root]
			if !exists {
				pr = &projRow{
					name: filepath.Base(root),
					path: shortenHomePath(root, home),
				}
				projMap[root] = pr
			}
			pr.b += it.Bytes
			continue
		}
		name := it.Name
		if it.ReportOnly || it.Mode == clean.ModeReportOnly {
			name = name + " (report-only)"
		}
		if len(name) > maxName {
			maxName = len(name)
		}
		rows = append(rows, row{name: name, b: it.Bytes})
		if it.Bytes > 0 {
			total += it.Bytes
		}
	}

	var projRows []projRow
	for _, pr := range projMap {
		projRows = append(projRows, *pr)
		if len(pr.name) > maxName {
			maxName = len(pr.name)
		}
		total += pr.b
	}

	sort.SliceStable(projRows, func(i, j int) bool { return projRows[i].b > projRows[j].b })
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].b > rows[j].b })

	for _, r := range projRows {
		size := "-"
		if r.b > 0 {
			size = humanBytes(r.b)
		}
		fmt.Fprintf(w, "%-*s  %s\n", maxName, r.name, size)
		fmt.Fprintf(w, "%s\n", r.path)
	}
	if len(projRows) > 0 && len(rows) > 0 {
		fmt.Fprintln(w)
	}

	for _, r := range rows {
		size := "-"
		if r.b > 0 {
			size = humanBytes(r.b)
		}
		fmt.Fprintf(w, "%-*s  %s\n", maxName, r.name, size)
	}
	if len(rows) > 0 || len(projRows) > 0 {
		fmt.Fprintf(w, "%s\n", strings.Repeat("-", maxName+2+12))
		fmt.Fprintf(w, "%-*s  %s\n", maxName, "Total", humanBytes(total))
	} else {
		fmt.Fprintln(w, "(no matching items found)")
	}
}

func projectRootFromID(id string) (string, bool) {
	idx := strings.Index(id, ":")
	if idx <= 0 {
		return "", false
	}
	prefix := id[:idx]
	if !strings.HasPrefix(prefix, "proj-") {
		return "", false
	}
	return id[idx+1:], true
}

func shortenHomePath(p, home string) string {
	if home == "" {
		return p
	}
	p = filepath.Clean(p)
	home = filepath.Clean(home)
	if p == home {
		return "~"
	}
	if strings.HasPrefix(p, home+string(filepath.Separator)) {
		return "~" + strings.TrimPrefix(p, home)
	}
	return p
}
