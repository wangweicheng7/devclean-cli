package clean

import (
	"time"
)

type Category string

const (
	CategoryCache Category = "cache"
	CategoryLogs  Category = "logs"
	CategoryBuild Category = "build"
)

type Profile string

const (
	ProfileSafe       Profile = "safe"
	ProfileDev        Profile = "dev"
	ProfileAggressive Profile = "aggressive"
)

type Mode string

const (
	ModeDelete     Mode = "delete"
	ModeReportOnly Mode = "report_only"
)

type Item struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Category    Category `json:"category"`
	ProfileMin  Profile  `json:"profile_min"`
	Mode        Mode     `json:"mode"`
	Exists      bool     `json:"exists"`
	Bytes       int64    `json:"bytes,omitempty"`
	FileCount   int64    `json:"file_count,omitempty"`
	Reason      string   `json:"reason,omitempty"`
	Warnings    []string `json:"warnings,omitempty"`
	ScannedAt   string   `json:"scanned_at,omitempty"`
	ReportOnly  bool     `json:"report_only,omitempty"`
	Skipped     bool     `json:"skipped,omitempty"`
	SkipReason  string   `json:"skip_reason,omitempty"`
	ResolvedAbs string   `json:"resolved_abs,omitempty"`
	AllowedRoot string   `json:"-"`
}

type Plan struct {
	Profile   Profile    `json:"profile"`
	Categories []Category `json:"categories"`
	Items     []Item     `json:"items"`
	TotalBytes int64     `json:"total_bytes,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

