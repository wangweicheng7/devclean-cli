package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/wangweicheng7/cleandev-cli/internal/clean"
)

func RunScan(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	fs.SetOutput(errOut)

	var (
		profile  = fs.String("profile", string(clean.ProfileSafe), "safe|dev|aggressive")
		category = fs.String("category", "", "comma-separated: cache,logs,build (default all)")
		repo     = fs.String("repo", "", "path to repository root (scan build artifacts; report-only)")
		asJSON   = fs.Bool("json", false, "output as json")
		withSize = fs.Bool("with-size", true, "calculate directory sizes (may be slow)")
	)

	if err := fs.Parse(args); err != nil {
		return 2
	}

	p, err := clean.ParseProfile(*profile)
	if err != nil {
		fmt.Fprintln(errOut, err.Error())
		return 2
	}

	cats, err := clean.ParseCategories(*category)
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
		WithSize:   *withSize,
		RepoRoot:   *repo,
	})
	if err != nil {
		fmt.Fprintln(errOut, err.Error())
		return 1
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

