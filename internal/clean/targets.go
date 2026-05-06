package clean

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Targets are intentionally conservative: only directories that are
// well-known to be safe-ish caches are deletable by default.
func DefaultTargets(home string) []Item {
	// Convenience for "~" expansions.
	h := func(p string) string {
		if strings.HasPrefix(p, "~/") {
			return filepath.Join(home, strings.TrimPrefix(p, "~/"))
		}
		return p
	}

	return []Item{
		{
			ID:         "go-build-cache",
			Name:       "Go build cache",
			Path:       h("~/Library/Caches/go-build"),
			Category:   CategoryCache,
			ProfileMin: ProfileSafe,
			Mode:       ModeDelete,
			Reason:     "rebuildable by `go build`",
		},
		{
			ID:         "npm-cache",
			Name:       "npm cache",
			Path:       h("~/Library/Caches/npm"),
			Category:   CategoryCache,
			ProfileMin: ProfileSafe,
			Mode:       ModeDelete,
			Reason:     "rebuildable by npm",
		},
		{
			ID:         "yarn-cache",
			Name:       "Yarn cache",
			Path:       h("~/Library/Caches/Yarn"),
			Category:   CategoryCache,
			ProfileMin: ProfileSafe,
			Mode:       ModeDelete,
			Reason:     "rebuildable by yarn",
		},
		{
			ID:         "pip-cache",
			Name:       "pip cache",
			Path:       h("~/Library/Caches/pip"),
			Category:   CategoryCache,
			ProfileMin: ProfileSafe,
			Mode:       ModeDelete,
			Reason:     "rebuildable by pip",
		},
		{
			ID:         "gradle-cache",
			Name:       "Gradle caches (user)",
			Path:       h("~/.gradle/caches"),
			Category:   CategoryCache,
			ProfileMin: ProfileDev,
			Mode:       ModeReportOnly,
			Reason:     "large rebuildable cache; default report-only to avoid surprising build changes",
			Warnings:   []string{"report-only (high churn cache)"},
			ReportOnly: true,
		},
		{
			ID:         "cocoapods-cache",
			Name:       "CocoaPods cache",
			Path:       h("~/Library/Caches/CocoaPods"),
			Category:   CategoryCache,
			ProfileMin: ProfileDev,
			Mode:       ModeDelete,
			Reason:     "rebuildable by CocoaPods",
		},
		{
			ID:         "npm-logs",
			Name:       "npm logs",
			Path:       h("~/.npm/_logs"),
			Category:   CategoryLogs,
			ProfileMin: ProfileDev,
			Mode:       ModeReportOnly,
			Reason:     "may contain useful debugging information; default report-only",
			Warnings:   []string{"report-only (optional)"},
			ReportOnly: true,
		},
		{
			ID:         "xcode-derived-data",
			Name:       "Xcode DerivedData",
			Path:       h("~/Library/Developer/Xcode/DerivedData"),
			Category:   CategoryBuild,
			ProfileMin: ProfileAggressive,
			Mode:       ModeReportOnly,
			Reason:     "large but can impact Xcode state; report-only by default",
			Warnings:   []string{"report-only (high impact)"},
			ReportOnly: true,
		},
		{
			ID:         "xcode-archives",
			Name:       "Xcode Archives",
			Path:       h("~/Library/Developer/Xcode/Archives"),
			Category:   CategoryBuild,
			ProfileMin: ProfileAggressive,
			Mode:       ModeReportOnly,
			Reason:     "archives may be needed for re-signing / distribution; report-only",
			Warnings:   []string{"report-only (keep if you need old archives)"},
			ReportOnly: true,
		},
	}
}

func ParseProfile(s string) (Profile, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", string(ProfileSafe):
		return ProfileSafe, nil
	case string(ProfileDev):
		return ProfileDev, nil
	case string(ProfileAggressive):
		return ProfileAggressive, nil
	default:
		return "", fmt.Errorf("invalid profile: %q (expected safe|dev|aggressive)", s)
	}
}

func ParseCategories(s string) ([]Category, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	var out []Category
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToLower(p))
		if p == "" {
			continue
		}
		switch Category(p) {
		case CategoryCache, CategoryLogs, CategoryBuild:
			out = append(out, Category(p))
		default:
			return nil, fmt.Errorf("invalid category: %q (expected cache|logs|build)", p)
		}
	}
	return out, nil
}

func HomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return home, nil
}

