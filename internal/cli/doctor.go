package cli

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
)

func RunDoctor(_ context.Context, _ []string, out io.Writer, _ io.Writer) int {
	fmt.Fprintf(out, "os: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	if runtime.GOOS != "darwin" {
		fmt.Fprintln(out, "warning: this tool currently targets macOS (darwin)")
	}

	if _, err := exec.LookPath("brew"); err != nil {
		fmt.Fprintln(out, "brew: not found (Homebrew not installed or not in PATH)")
	} else {
		fmt.Fprintln(out, "brew: found")
	}

	return 0
}

