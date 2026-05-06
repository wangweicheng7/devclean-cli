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
		fmt.Fprintln(errOut, "usage: devclean config init|prune-missing ...")
		return 2
	}

	switch args[0] {
	case "init":
		return runConfigInit(args[1:], out, errOut)
	case "prune-missing":
		return runConfigPruneMissing(args[1:], out, errOut)
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

func runConfigPruneMissing(args []string, out io.Writer, errOut io.Writer) int {
	fs := flag.NewFlagSet("config prune-missing", flag.ContinueOnError)
	fs.SetOutput(errOut)

	configPath := fs.String("config", "", "path to config file (default: .devcleanrc.json in current dir)")
	apply := fs.Bool("apply", false, "write changes back to config file")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	cfg, usedPath, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintln(errOut, validateConfigPath(err, usedPath).Error())
		return 1
	}
	if usedPath == "" {
		fmt.Fprintln(errOut, "no config file found (use --config or run `devclean config init` first)")
		return 1
	}

	removedInc, keptInc := pruneMissingProjectIDs(cfg.IncludeIDs)
	removedExc, keptExc := pruneMissingProjectIDs(cfg.ExcludeIDs)

	if len(removedInc) == 0 && len(removedExc) == 0 {
		fmt.Fprintln(out, "no missing project ids to prune")
		return 0
	}

	for _, id := range removedInc {
		fmt.Fprintf(out, "would remove include_id: %s\n", id)
	}
	for _, id := range removedExc {
		fmt.Fprintf(out, "would remove exclude_id: %s\n", id)
	}

	if !*apply {
		fmt.Fprintln(out, "dry-run: pass --apply to write changes")
		return 0
	}

	cfg.IncludeIDs = keptInc
	cfg.ExcludeIDs = keptExc
	if err := config.Save(usedPath, cfg); err != nil {
		fmt.Fprintln(errOut, err.Error())
		return 1
	}
	fmt.Fprintf(out, "updated config: %s\n", usedPath)
	return 0
}


