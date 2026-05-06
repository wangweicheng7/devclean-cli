package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
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

	printPlanText(out, plan)
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

func printPlanText(w io.Writer, plan clean.Plan) {
	fmt.Fprintf(w, "profile: %s\n", plan.Profile)
	if len(plan.Categories) > 0 {
		var parts []string
		for _, c := range plan.Categories {
			parts = append(parts, string(c))
		}
		fmt.Fprintf(w, "categories: %s\n", strings.Join(parts, ","))
	}
	fmt.Fprintln(w)

	for _, it := range plan.Items {
		status := "missing"
		if it.Skipped {
			status = "skipped"
		} else if it.Exists {
			status = "found"
		}
		mode := string(it.Mode)
		if it.ReportOnly {
			mode = "report_only"
		}
		size := ""
		if it.Bytes > 0 {
			size = fmt.Sprintf(" (%s)", humanBytes(it.Bytes))
		}
		fmt.Fprintf(w, "- [%s] %s: %s%s\n", status, it.Name, mode, size)
		fmt.Fprintf(w, "  path: %s\n", it.Path)
		if it.Reason != "" {
			fmt.Fprintf(w, "  reason: %s\n", it.Reason)
		}
		if it.SkipReason != "" {
			fmt.Fprintf(w, "  skip: %s\n", it.SkipReason)
		}
		if len(it.Warnings) > 0 {
			fmt.Fprintf(w, "  warnings: %s\n", strings.Join(it.Warnings, "; "))
		}
	}
}

