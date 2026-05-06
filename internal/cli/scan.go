package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/wangweicheng7/devclean-cli/internal/clean"
)

func RunScan(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	fs.SetOutput(errOut)

	configPath := &stringFlag{}
	fs.Var(configPath, "config", "path to config file (default: .devcleanrc.json in current dir)")

	profile := &stringFlag{v: string(clean.ProfileSafe)}
	fs.Var(profile, "profile", "safe|dev|aggressive")

	category := &stringFlag{}
	fs.Var(category, "category", "comma-separated: cache,logs,build (default all)")

	repo := &stringFlag{}
	fs.Var(repo, "repo", "path to repository root (scan build artifacts; report-only)")

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

	p, err := clean.ParseProfile(profile.v)
	if err != nil {
		fmt.Fprintln(errOut, err.Error())
		return 2
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

	plan, err := clean.BuildPlan(ctx, clean.ScanOptions{
		Profile:    p,
		Categories: catSet,
		WithSize:   withSizeFlag.v,
		RepoRoot:   repo.v,
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

func printPlanTable(w io.Writer, plan clean.Plan) {
	// Collect items that actually exist on disk. This keeps output focused.
	type row struct {
		name string
		b    int64
	}
	var rows []row
	var total int64
	maxName := 0

	for _, it := range plan.Items {
		if !it.Exists || it.Skipped {
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

	sort.SliceStable(rows, func(i, j int) bool { return rows[i].b > rows[j].b })

	for _, r := range rows {
		size := "-"
		if r.b > 0 {
			size = humanBytes(r.b)
		}
		fmt.Fprintf(w, "%-*s  %s\n", maxName, r.name, size)
	}
	if len(rows) > 0 {
		fmt.Fprintf(w, "%s\n", strings.Repeat("-", maxName+2+12))
		fmt.Fprintf(w, "%-*s  %s\n", maxName, "Total", humanBytes(total))
	} else {
		fmt.Fprintln(w, "(no matching items found)")
	}
}

