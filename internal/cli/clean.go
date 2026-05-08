package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wangweicheng7/devclean-cli/internal/clean"
)

func RunClean(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	fs := flag.NewFlagSet("clean", flag.ContinueOnError)
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

	dryRun := fs.Bool("dry-run", false, "preview actions without deleting")
	confirm := fs.Bool("confirm", false, "execute deletion (required unless --dry-run)")
	allowReportOnly := fs.Bool("allow-report-only", false, "allow deleting report-only items (Xcode/Gradle/User caches). Use with care.")
	interactive := fs.Bool("interactive", false, "interactively confirm each candidate item")
	interactiveBatch := fs.Bool("interactive-batch", false, "batch mode: choose items first, then execute at the end (legacy behavior)")
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
	if *interactive && *asJSON {
		fmt.Fprintln(errOut, "--interactive cannot be combined with --json")
		return 2
	}

	var selectedIDs []string
	var discoverLogs []string
	if *interactive {
		stopSpinner := startSpinner(errOut, "scanning")
		defer stopSpinner()

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
		stopSpinner()
		if err != nil {
			fmt.Fprintln(errOut, err.Error())
			return 1
		}
		if len(cfg.IncludeIDs) > 0 || len(cfg.ExcludeIDs) > 0 {
			plan.Items = filterItemsByIDs(plan.Items, cfg.IncludeIDs, cfg.ExcludeIDs)
		}
		applyImmediately := !*interactiveBatch
		ids, err := chooseInteractive(ctx, plan, *allowReportOnly, applyImmediately, *dryRun, out, errOut)
		if err != nil {
			fmt.Fprintln(errOut, err.Error())
			return 1
		}
		// When applying immediately, deletion is already done item-by-item.
		if applyImmediately {
			return 0
		}
		if len(ids) == 0 {
			fmt.Fprintln(out, "interactive: no item selected, nothing to do")
			return 0
		}
		selectedIDs = ids
	}

	// In interactive mode, user's per-item Y/N is already an explicit authorization,
	// so we don't require a separate `--confirm` flag.
	confirmForExecute := *confirm || *interactive

	res, err := clean.Execute(ctx, clean.ExecuteOptions{
		DryRun:          *dryRun,
		Confirm:         confirmForExecute,
		Profile:         p,
		Category:        catSet,
		WithSize:        withSizeFlag.v,
		All:             *all,
		RepoRoot:        repo.v,
		UserCaches:      *userCaches,
		TargetIDs:       selectedIDs,
		ExcludeIDs:      cfg.ExcludeIDs,
		AllowReportOnly: *allowReportOnly,
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

	if *asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
		return 0
	}
	if *discoverDebug {
		for _, l := range discoverLogs {
			fmt.Fprintf(errOut, "[discover] %s\n", l)
		}
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

func chooseInteractive(ctx context.Context, plan clean.Plan, allowReportOnly bool, applyImmediately bool, dryRun bool, out io.Writer, errOut io.Writer) ([]string, error) {
	reader := bufio.NewReader(os.Stdin)
	var selected []string
	var applied int
	projectGroups := map[string][]clean.Item{}
	var projectRootsOrdered []string
	var singles []clean.Item

	for _, it := range plan.Items {
		if !it.Exists || it.Skipped {
			continue
		}
		if (it.ReportOnly || it.Mode == clean.ModeReportOnly) && !allowReportOnly {
			continue
		}
		if it.Mode != clean.ModeDelete && !(allowReportOnly && (it.ReportOnly || it.Mode == clean.ModeReportOnly)) {
			continue
		}
		// Reduce confirmation noise: skip empty directories when size is known.
		// (When WithSize is disabled, Bytes/FileCount will be 0 even for non-empty dirs,
		// so this only triggers when file_count is actually computed as 0.)
		if it.Bytes == 0 && it.FileCount == 0 && !hasSizeFailure(it.Warnings) {
			continue
		}
		if root, ok := projectRootFromID(it.ID); ok && root != "" {
			if _, exists := projectGroups[root]; !exists {
				projectRootsOrdered = append(projectRootsOrdered, root)
			}
			projectGroups[root] = append(projectGroups[root], it)
			continue
		}
		singles = append(singles, it)
	}

	// Project-first interaction: one confirmation per project, delete selected
	// project's matched cache/build dirs together.
	for _, root := range projectRootsOrdered {
		items := projectGroups[root]
		if len(items) == 0 {
			continue
		}
		projectName := filepath.Base(root)
		var totalBytes int64
		for _, it := range items {
			totalBytes += it.Bytes
		}
		totalSize := ""
		if totalBytes > 0 {
			totalSize = fmt.Sprintf(" (%s)", humanBytes(totalBytes))
		}
		fmt.Fprintf(out, "clean project? %s%s\n  root: %s\n", projectName, totalSize, root)
		for _, it := range items {
			path := it.ResolvedAbs
			if path == "" {
				path = it.Path
			}
			size := "-"
			if it.Bytes > 0 {
				size = humanBytes(it.Bytes)
			}
			// Keep project item list concise and scannable.
			fmt.Fprintf(out, "  - %s  %s\n", it.Name, size)
			fmt.Fprintf(out, "    %s\n", path)
		}
		fmt.Fprint(out, "[y/N]: ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		answer := strings.ToLower(strings.TrimSpace(line))
		if answer != "y" && answer != "yes" {
			continue
		}

		for _, it := range items {
			path := it.ResolvedAbs
			if path == "" {
				path = it.Path
			}
			if applyImmediately {
				ok, err := clean.ExecuteItem(ctx, it, clean.ExecuteItemOptions{
					DryRun:          dryRun,
					AllowReportOnly: allowReportOnly,
				})
				if err != nil {
					return nil, err
				}
				if ok {
					applied++
					if dryRun {
						fmt.Fprintf(out, "would delete: %s\n", path)
					} else {
						fmt.Fprintf(out, "deleted: %s\n", path)
					}
				} else {
					fmt.Fprintf(out, "skipped: %s\n", path)
				}
				continue
			}
			selected = append(selected, it.ID)
		}
		fmt.Fprintln(out)
	}

	for _, it := range singles {
		size := ""
		if it.Bytes > 0 {
			size = fmt.Sprintf(" (%s)", humanBytes(it.Bytes))
		}
		path := it.ResolvedAbs
		if path == "" {
			path = it.Path
		}
		fmt.Fprintf(out, "clean item? %s%s\n  path: %s\n[y/N]: ", it.Name, size, path)
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		answer := strings.ToLower(strings.TrimSpace(line))
		if answer == "y" || answer == "yes" {
			if applyImmediately {
				ok, err := clean.ExecuteItem(ctx, it, clean.ExecuteItemOptions{
					DryRun:          dryRun,
					AllowReportOnly: allowReportOnly,
				})
				if err != nil {
					return nil, err
				}
				if ok {
					applied++
					if dryRun {
						fmt.Fprintf(out, "would delete: %s\n\n", path)
					} else {
						fmt.Fprintf(out, "deleted: %s\n\n", path)
					}
				} else {
					fmt.Fprintf(out, "skipped: %s\n\n", path)
				}
				continue
			}
			selected = append(selected, it.ID)
		}
	}
	if applyImmediately {
		if applied == 0 {
			fmt.Fprintln(errOut, "interactive: no executable items selected")
		} else if dryRun {
			fmt.Fprintf(errOut, "interactive: would delete %d item(s)\n", applied)
		} else {
			fmt.Fprintf(errOut, "interactive: deleted %d item(s)\n", applied)
		}
		return nil, nil
	}
	if len(selected) == 0 {
		fmt.Fprintln(errOut, "interactive: no executable items selected")
	}
	return selected, nil
}

func hasSizeFailure(warnings []string) bool {
	for _, w := range warnings {
		if strings.Contains(strings.ToLower(w), "size calculation failed") {
			return true
		}
	}
	return false
}

func startSpinner(w io.Writer, label string) func() {
	done := make(chan struct{})
	go func() {
		frames := []string{"|", "/", "-", "\\"}
		i := 0
		for {
			select {
			case <-done:
				// clear line
				fmt.Fprintf(w, "\r%s... done%s\r", label, strings.Repeat(" ", 20))
				return
			case <-time.After(120 * time.Millisecond):
				fmt.Fprintf(w, "\r%s... %s", label, frames[i%len(frames)])
				i++
			}
		}
	}()
	var once bool
	return func() {
		if once {
			return
		}
		once = true
		close(done)
	}
}
