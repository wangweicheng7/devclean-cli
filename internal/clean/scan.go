package clean

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ScanOptions struct {
	Profile    Profile
	Categories map[Category]bool // nil or empty => all
	WithSize   bool
	RepoRoot   string // when set, additionally scan repo build artifacts (report-only)
	Discover   DiscoverOptions
	UserCaches bool // include ~/Library/Caches/* (top-level only, report-only)
}

func BuildPlan(ctx context.Context, opts ScanOptions) (Plan, error) {
	home, err := HomeDir()
	if err != nil {
		return Plan{}, err
	}

	all := DefaultTargets(home)
	var items []Item
	addItems := func(root string, targetItems []Item) error {
		for _, it := range targetItems {
			if !profileAllows(opts.Profile, it.ProfileMin) {
				continue
			}
			if len(opts.Categories) > 0 && !opts.Categories[it.Category] {
				continue
			}

			scanned := it
			scanned.ScannedAt = time.Now().Format(time.RFC3339)

			abs, resolved, warn, skip, skipReason := resolveAndValidatePathAllowedRoot(root, it.Path)
			scanned.ResolvedAbs = resolved
			scanned.AllowedRoot = root
			if warn != "" {
				scanned.Warnings = append(scanned.Warnings, warn)
			}
			if skip {
				scanned.Skipped = true
				scanned.SkipReason = skipReason
				items = append(items, scanned)
				continue
			}

			st, err := os.Stat(abs)
			if err != nil {
				if os.IsNotExist(err) {
					scanned.Exists = false
					items = append(items, scanned)
					continue
				}
				return fmt.Errorf("stat %s: %w", abs, err)
			}
			if !st.IsDir() {
				scanned.Skipped = true
				scanned.SkipReason = "not a directory"
				items = append(items, scanned)
				continue
			}

			scanned.Exists = true

			if opts.WithSize {
				bytes, files, err := dirSize(ctx, abs)
				if err != nil {
					scanned.Warnings = append(scanned.Warnings, "size calculation failed: "+err.Error())
				} else {
					scanned.Bytes = bytes
					scanned.FileCount = files
				}
			}

			items = append(items, scanned)
		}
		return nil
	}

	// Home targets.
	if err := addItems(home, all); err != nil {
		return Plan{}, err
	}

	// User Library caches (report-only).
	if opts.UserCaches {
		if err := addItems(home, UserLibraryCachesTargets(home)); err != nil {
			return Plan{}, err
		}
	}

	// Repo targets (report-only).
	if strings.TrimSpace(opts.RepoRoot) != "" {
		repoResolved, err := resolveRepoRoot(opts.RepoRoot)
		if err != nil {
			return Plan{}, err
		}
		if err := addItems(repoResolved, RepoTargets(repoResolved)); err != nil {
			return Plan{}, err
		}
	}

	// Auto-discovered project targets.
	if opts.Discover.Enabled {
		projects, err := discoverProjects(ctx, home, opts.Discover)
		if err != nil {
			return Plan{}, err
		}
		for _, p := range projects {
			if err := addItems(p.Root, projectTargets(p)); err != nil {
				return Plan{}, err
			}
		}
	}

	var cats []Category
	if len(opts.Categories) == 0 {
		cats = []Category{CategoryCache, CategoryLogs, CategoryBuild}
	} else {
		for c := range opts.Categories {
			cats = append(cats, c)
		}
	}

	p := Plan{
		Profile:    opts.Profile,
		Categories: cats,
		Items:      items,
		CreatedAt:  time.Now(),
	}
	for _, it := range items {
		if it.Exists && !it.Skipped && it.Mode == ModeDelete && it.Bytes > 0 {
			p.TotalBytes += it.Bytes
		}
	}
	return p, nil
}

func resolveRepoRoot(repoRoot string) (resolvedRoot string, err error) {
	abs, err := filepath.Abs(repoRoot)
	if err != nil {
		return "", fmt.Errorf("invalid --repo path: %w", err)
	}
	abs = filepath.Clean(abs)

	if pathHasGitSegment(abs) {
		return "", fmt.Errorf("refusing to scan a repo root under .git: %s", abs)
	}

	// Validate it is a directory (even if we might not have permission to read inside).
	st, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("--repo not accessible: %w", err)
	}
	if !st.IsDir() {
		return "", fmt.Errorf("--repo is not a directory: %s", abs)
	}

	resolvedRoot = abs
	if _, err := os.Lstat(abs); err == nil {
		if r, err := filepath.EvalSymlinks(abs); err == nil {
			resolvedRoot = r
		}
	}
	if pathHasGitSegment(resolvedRoot) {
		return "", fmt.Errorf("refusing to scan a repo root under .git: %s", resolvedRoot)
	}
	return resolvedRoot, nil
}

func profileAllows(selected, min Profile) bool {
	order := map[Profile]int{
		ProfileSafe:       1,
		ProfileDev:        2,
		ProfileAggressive: 3,
	}
	return order[selected] >= order[min]
}

func pathHasGitSegment(p string) bool {
	// Use slash normalization so it works across platforms.
	parts := strings.Split(filepath.ToSlash(filepath.Clean(p)), "/")
	for _, part := range parts {
		if part == ".git" {
			return true
		}
	}
	return false
}

func resolveAndValidatePathAllowedRoot(allowedRoot, p string) (abs string, resolved string, warning string, skip bool, skipReason string) {
	abs = p
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(allowedRoot, p)
	}
	abs = filepath.Clean(abs)

	// Hard protection rules.
	if abs == "/" || abs == allowedRoot || abs == filepath.Clean(allowedRoot+string(filepath.Separator)) {
		return abs, abs, "", true, "protected root"
	}
	if pathHasGitSegment(abs) {
		return abs, abs, "", true, "protected .git path"
	}
	if !filepath.IsAbs(abs) {
		return abs, abs, "", true, "not absolute after clean"
	}

	// Resolve symlinks if it exists; if not exist, keep cleaned path.
	resolved = abs
	if _, err := os.Lstat(abs); err == nil {
		if r, err := filepath.EvalSymlinks(abs); err == nil {
			resolved = r
		} else {
			warning = "failed to resolve symlinks"
		}
	}
	if pathHasGitSegment(resolved) {
		return abs, resolved, "", true, "protected .git path (via symlink)"
	}

	// Ensure within allowed root. This is a conservative safety net.
	rel, err := filepath.Rel(allowedRoot, resolved)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || strings.HasPrefix(rel, ".."+"/") {
		return abs, resolved, "", true, "outside allowed root directory"
	}

	return abs, resolved, warning, false, ""
}

func dirSize(ctx context.Context, root string) (bytes int64, files int64, err error) {
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		bytes += info.Size()
		files++
		return nil
	})
	return bytes, files, err
}
