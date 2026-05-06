package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/wangweicheng7/devclean-cli/internal/config"
)

func RunConfig(_ context.Context, args []string, out io.Writer, errOut io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(errOut, "usage: devclean config init [--path .devcleanrc.json] [--force]")
		return 2
	}

	switch args[0] {
	case "init":
		return runConfigInit(args[1:], out, errOut)
	default:
		fmt.Fprintf(errOut, "unknown config subcommand: %s\n", args[0])
		return 2
	}
}

func runConfigInit(args []string, out io.Writer, errOut io.Writer) int {
	fs := flag.NewFlagSet("config init", flag.ContinueOnError)
	fs.SetOutput(errOut)

	path := fs.String("path", config.DefaultConfigFilename, "path to write config template")
	force := fs.Bool("force", false, "overwrite existing file")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	p := *path
	if !filepath.IsAbs(p) {
		cwd, _ := os.Getwd()
		p = filepath.Join(cwd, p)
	}

	if err := config.WriteTemplate(p, *force); err != nil {
		fmt.Fprintln(errOut, err.Error())
		return 1
	}

	fmt.Fprintf(out, "wrote config template: %s\n", p)
	return 0
}

