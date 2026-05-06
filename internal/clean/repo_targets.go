package clean

import "path/filepath"

// RepoTargets returns repo build artifact candidates.
// They are report-only by default to keep the tool safe in the presence of project-specific semantics.
func RepoTargets(repoRoot string) []Item {
	return []Item{
		{
			ID:         "repo-build",
			Name:       "Repository build/",
			Path:       filepath.Join(repoRoot, "build"),
			Category:   CategoryBuild,
			ProfileMin: ProfileDev,
			Mode:       ModeReportOnly,
			Reason:     "build artifact; report-only by default",
			ReportOnly: true,
			Warnings:   []string{"report-only (MVP)" },
		},
		{
			ID:         "repo-dist",
			Name:       "Repository dist/",
			Path:       filepath.Join(repoRoot, "dist"),
			Category:   CategoryBuild,
			ProfileMin: ProfileDev,
			Mode:       ModeReportOnly,
			Reason:     "build artifact; report-only by default",
			ReportOnly: true,
			Warnings:   []string{"report-only (MVP)" },
		},
		{
			ID:         "repo-target",
			Name:       "Repository target/",
			Path:       filepath.Join(repoRoot, "target"),
			Category:   CategoryBuild,
			ProfileMin: ProfileDev,
			Mode:       ModeReportOnly,
			Reason:     "build artifact; report-only by default",
			ReportOnly: true,
			Warnings:   []string{"report-only (MVP)" },
		},
		{
			ID:         "repo-out",
			Name:       "Repository out/",
			Path:       filepath.Join(repoRoot, "out"),
			Category:   CategoryBuild,
			ProfileMin: ProfileDev,
			Mode:       ModeReportOnly,
			Reason:     "build artifact; report-only by default",
			ReportOnly: true,
			Warnings:   []string{"report-only (MVP)" },
		},
		{
			ID:         "repo-gradle-cache",
			Name:       "Repository .gradle/",
			Path:       filepath.Join(repoRoot, ".gradle"),
			Category:   CategoryBuild,
			ProfileMin: ProfileDev,
			Mode:       ModeReportOnly,
			Reason:     "Gradle local state; report-only by default",
			ReportOnly: true,
			Warnings:   []string{"report-only (MVP)" },
		},
		{
			ID:         "repo-node_modules",
			Name:       "Repository node_modules/",
			Path:       filepath.Join(repoRoot, "node_modules"),
			Category:   CategoryBuild,
			ProfileMin: ProfileDev,
			Mode:       ModeReportOnly,
			Reason:     "dependency install output; report-only by default",
			ReportOnly: true,
			Warnings:   []string{"report-only (MVP)", "may be large"} ,
		},
		{
			ID:         "repo-pods",
			Name:       "Repository Pods/",
			Path:       filepath.Join(repoRoot, "Pods"),
			Category:   CategoryBuild,
			ProfileMin: ProfileDev,
			Mode:       ModeReportOnly,
			Reason:     "CocoaPods install output; report-only by default",
			ReportOnly: true,
			Warnings:   []string{"report-only (MVP)"} ,
		},
	}
}

