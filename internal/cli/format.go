package cli

import "fmt"

func humanBytes(n int64) string {
	// Human-friendly, decimal units (MB/GB) as typically expected by users.
	if n < 1000 {
		return fmt.Sprintf("%d B", n)
	}
	const unit = 1000
	div, exp := float64(unit), 0
	v := float64(n)
	for v/div >= unit && exp < 3 {
		div *= unit
		exp++
	}
	suffix := []string{"KB", "MB", "GB", "TB"}[exp]
	return fmt.Sprintf("%.2f %s", float64(n)/div, suffix)
}

