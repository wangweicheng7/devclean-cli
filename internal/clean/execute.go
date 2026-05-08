package clean

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type ExecuteOptions struct {
	DryRun          bool
	Confirm         bool
	Profile         Profile
	Category        map[Category]bool // nil or empty => all
	WithSize        bool
	RepoRoot        string
	TargetIDs       []string // nil/empty => all executable items
	ExcludeIDs      []string // ids to always skip
	Discover        DiscoverOptions
	UserCaches      bool
	AllowReportOnly bool // allow deletion of report-only items (explicit opt-in)
}

type ExecuteResult struct {
	Plan       Plan     `json:"plan"`
	DeletedIDs []string `json:"deleted_ids,omitempty"`
	SkippedIDs []string `json:"skipped_ids,omitempty"`
	DryRun     bool     `json:"dry_run"`
}

type ExecuteItemOptions struct {
	DryRun          bool
	AllowReportOnly bool
}

// ExecuteItem executes deletion for a single already-planned item, applying the same
// safety rules as Execute(). It is useful for interactive "apply immediately" flows.
func ExecuteItem(ctx context.Context, it Item, opts ExecuteItemOptions) (deleted bool, err error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	home, err := HomeDir()
	if err != nil {
		return false, err
	}

	if !it.Exists || it.Skipped {
		return false, nil
	}

	// By default we never delete report-only items. Only allow when explicitly requested.
	if it.ReportOnly || it.Mode == ModeReportOnly {
		if !opts.AllowReportOnly {
			return false, nil
		}
	} else if it.Mode != ModeDelete {
		return false, nil
	}

	// Prefer plan-resolved absolute path (helps avoid symlink race conditions).
	candidate := it.ResolvedAbs
	if candidate == "" {
		candidate = it.Path
		if !filepath.IsAbs(candidate) {
			candidate = filepath.Join(home, candidate)
		}
	}
	candidate = filepath.Clean(candidate)

	allowedRoot := it.AllowedRoot
	if allowedRoot == "" {
		allowedRoot = home
	}
	_, _, _, skip, _ := resolveAndValidatePathAllowedRoot(allowedRoot, candidate)
	if skip {
		return false, nil
	}

	if opts.DryRun {
		return true, nil
	}
	if err := os.RemoveAll(candidate); err != nil {
		return false, fmt.Errorf("remove %s: %w", candidate, err)
	}
	return true, nil
}

func Execute(ctx context.Context, opts ExecuteOptions) (ExecuteResult, error) {
	// If both are set, we treat it as dry-run to avoid surprising deletions.
	if opts.DryRun && opts.Confirm {
		opts.Confirm = false
	}
	if !opts.DryRun && !opts.Confirm {
		return ExecuteResult{}, fmt.Errorf("refusing to clean without --confirm (or use --dry-run to preview)")
	}

	p, err := BuildPlan(ctx, ScanOptions{
		Profile:    opts.Profile,
		Categories: opts.Category,
		WithSize:   opts.WithSize,
		RepoRoot:   opts.RepoRoot,
		Discover:   opts.Discover,
		UserCaches: opts.UserCaches,
	})
	if err != nil {
		return ExecuteResult{}, err
	}

	res := ExecuteResult{Plan: p, DryRun: opts.DryRun}
	targetSet := map[string]bool{}
	for _, id := range opts.TargetIDs {
		targetSet[id] = true
	}
	excludeSet := map[string]bool{}
	for _, id := range opts.ExcludeIDs {
		excludeSet[id] = true
	}

	home, err := HomeDir()
	if err != nil {
		return ExecuteResult{}, err
	}

	for _, it := range p.Items {
		if excludeSet[it.ID] {
			res.SkippedIDs = append(res.SkippedIDs, it.ID)
			continue
		}
		if len(targetSet) > 0 && !targetSet[it.ID] {
			res.SkippedIDs = append(res.SkippedIDs, it.ID)
			continue
		}
		if !it.Exists || it.Skipped {
			res.SkippedIDs = append(res.SkippedIDs, it.ID)
			continue
		}
		// By default we never delete report-only items. Only allow when explicitly requested.
		if it.ReportOnly || it.Mode == ModeReportOnly {
			if !opts.AllowReportOnly {
				res.SkippedIDs = append(res.SkippedIDs, it.ID)
				continue
			}
		} else if it.Mode != ModeDelete {
			res.SkippedIDs = append(res.SkippedIDs, it.ID)
			continue
		}

		// Prefer plan-resolved absolute path (helps avoid symlink race conditions).
		candidate := it.ResolvedAbs
		if candidate == "" {
			candidate = it.Path
			if !filepath.IsAbs(candidate) {
				candidate = filepath.Join(home, candidate)
			}
		}
		candidate = filepath.Clean(candidate)

		allowedRoot := it.AllowedRoot
		if allowedRoot == "" {
			allowedRoot = home
		}
		_, _, _, skip, _ := resolveAndValidatePathAllowedRoot(allowedRoot, candidate)
		if skip {
			res.SkippedIDs = append(res.SkippedIDs, it.ID)
			continue
		}

		if opts.DryRun {
			res.DeletedIDs = append(res.DeletedIDs, it.ID)
			continue
		}

		if err := os.RemoveAll(candidate); err != nil {
			return ExecuteResult{}, fmt.Errorf("remove %s: %w", candidate, err)
		}
		res.DeletedIDs = append(res.DeletedIDs, it.ID)
	}

	return res, nil
}
