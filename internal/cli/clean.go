package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/wangweicheng7/devclean-cli/internal/clean"
)

func RunClean(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	fs := flag.NewFlagSet("clean", flag.ContinueOnError)
	fs.SetOutput(errOut)

	configPath := &stringFlag{}
	fs.Var(configPath, "config", "path to config file (default: .devcleanrc.json in current dir)")

	profile := &stringFlag{v: string(clean.ProfileSafe)}
	fs.Var(profile, "profile", "safe|dev|aggressive")

	category := &stringFlag{}
	fs.Var(category, "category", "comma-separated: cache,logs,build (default all)")

	repo := &stringFlag{}
	fs.Var(repo, "repo", "path to repository root (scan build artifacts; report-only)")

	dryRun := fs.Bool("dry-run", false, "preview actions without deleting")
	confirm := fs.Bool("confirm", false, "execute deletion (required unless --dry-run)")
	interactive := fs.Bool("interactive", false, "interactively confirm each candidate item")
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
	if *interactive && *asJSON {
		fmt.Fprintln(errOut, "--interactive cannot be combined with --json")
		return 2
	}

	var selectedIDs []string
	if *interactive {
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
		ids, err := chooseInteractive(plan, out, errOut)
		if err != nil {
			fmt.Fprintln(errOut, err.Error())
			return 1
		}
		if len(ids) == 0 {
			fmt.Fprintln(out, "interactive: no item selected, nothing to do")
			return 0
		}
		selectedIDs = ids
	}

	res, err := clean.Execute(ctx, clean.ExecuteOptions{
		DryRun:    *dryRun,
		Confirm:   *confirm,
		Profile:   p,
		Category:  catSet,
		WithSize:  withSizeFlag.v,
		RepoRoot:  repo.v,
		TargetIDs: selectedIDs,
		ExcludeIDs: cfg.ExcludeIDs,
	})
	if err != nil {
		fmt.Fprintln(errOut, err.Error())
		return 1
	}

	if *asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
		return 0
	}

	if res.DryRun {
		fmt.Fprintln(out, "dry-run: no files were deleted")
		if len(res.DeletedIDs) > 0 {
			fmt.Fprintln(out, "would delete:")
			for _, id := range res.DeletedIDs {
				fmt.Fprintf(out, "  - %s\n", id)
			}
		} else {
			fmt.Fprintln(out, "would delete: (none)")
		}
	} else {
		fmt.Fprintln(out, "clean: deletion executed")
		if len(res.DeletedIDs) > 0 {
			fmt.Fprintln(out, "deleted:")
			for _, id := range res.DeletedIDs {
				fmt.Fprintf(out, "  - %s\n", id)
			}
		} else {
			fmt.Fprintln(out, "deleted: (none)")
		}
	}
	printPlanTable(out, res.Plan)
	return 0
}

func chooseInteractive(plan clean.Plan, out io.Writer, errOut io.Writer) ([]string, error) {
	reader := bufio.NewReader(os.Stdin)
	var selected []string
	for _, it := range plan.Items {
		if !it.Exists || it.Skipped || it.Mode != clean.ModeDelete || it.ReportOnly {
			continue
		}
		size := ""
		if it.Bytes > 0 {
			size = fmt.Sprintf(" (%s)", humanBytes(it.Bytes))
		}
		fmt.Fprintf(out, "clean item? %s%s [y/N]: ", it.Name, size)
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		answer := strings.ToLower(strings.TrimSpace(line))
		if answer == "y" || answer == "yes" {
			selected = append(selected, it.ID)
		}
	}
	if len(selected) == 0 {
		fmt.Fprintln(errOut, "interactive: no executable items selected")
	}
	return selected, nil
}

