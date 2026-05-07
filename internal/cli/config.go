package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/wangweicheng7/devclean-cli/internal/config"
)

func RunConfig(_ context.Context, args []string, out io.Writer, errOut io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(errOut, "usage: devclean config init|prune-missing|exclude|include ...")
		return 2
	}

	switch args[0] {
	case "init":
		return runConfigInit(args[1:], out, errOut)
	case "prune-missing":
		return runConfigPruneMissing(args[1:], out, errOut)
	case "exclude":
		return runConfigListEdit(args[1:], out, errOut, "exclude")
	case "include":
		return runConfigListEdit(args[1:], out, errOut, "include")
	default:
		fmt.Fprintf(errOut, "unknown config subcommand: %s\n", args[0])
		return 2
	}
}

func runConfigInit(args []string, out io.Writer, errOut io.Writer) int {
	fs := flag.NewFlagSet("config init", flag.ContinueOnError)
	fs.SetOutput(errOut)

	path := fs.String("path", "", "path to write config template (default: ~/.devcleanrc.json)")
	force := fs.Bool("force", false, "overwrite existing file")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	p := *path
	if p == "" {
		home, _ := os.UserHomeDir()
		if home == "" {
			fmt.Fprintln(errOut, "cannot determine home directory (pass --path explicitly)")
			return 1
		}
		p = filepath.Join(home, config.DefaultConfigFilename)
	} else if !filepath.IsAbs(p) {
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

	configPath := fs.String("config", "", "path to config file (default: .devcleanrc.json in current dir; fallback: ~/.devcleanrc.json)")
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

func runConfigListEdit(args []string, out io.Writer, errOut io.Writer, which string) int {
	if len(args) == 0 {
		fmt.Fprintf(errOut, "usage: devclean config %s add|remove|list ...\n", which)
		return 2
	}

	switch args[0] {
	case "list":
		return runConfigList(args[1:], out, errOut, which)
	case "add":
		return runConfigAddRemove(args[1:], out, errOut, which, true)
	case "remove":
		return runConfigAddRemove(args[1:], out, errOut, which, false)
	default:
		fmt.Fprintf(errOut, "unknown config %s subcommand: %s\n", which, args[0])
		return 2
	}
}

func runConfigList(args []string, out io.Writer, errOut io.Writer, which string) int {
	configPath, _, ids, err := parseConfigEditArgs(args)
	if err != nil {
		fmt.Fprintln(errOut, err.Error())
		return 2
	}
	if len(ids) > 0 {
		fmt.Fprintln(errOut, "usage: devclean config "+which+" list [--config path]")
		return 2
	}

	cfg, usedPath, err := loadConfig(configPath)
	if err != nil {
		fmt.Fprintln(errOut, validateConfigPath(err, usedPath).Error())
		return 1
	}
	if usedPath == "" {
		fmt.Fprintln(errOut, "no config file found (use --config or run `devclean config init` first)")
		return 1
	}

	var listIDs []string
	if which == "exclude" {
		listIDs = cfg.ExcludeIDs
	} else {
		listIDs = cfg.IncludeIDs
	}
	fmt.Fprintf(out, "config: %s\n", usedPath)
	if len(listIDs) == 0 {
		fmt.Fprintf(out, "%s_ids: (empty)\n", which)
		return 0
	}
	for _, id := range listIDs {
		fmt.Fprintf(out, "- %s\n", id)
	}
	return 0
}

func runConfigAddRemove(args []string, out io.Writer, errOut io.Writer, which string, add bool) int {
	configPath, dryRun, ids, err := parseConfigEditArgs(args)
	if err != nil {
		fmt.Fprintln(errOut, err.Error())
		return 2
	}
	if len(ids) == 0 {
		action := "add"
		if !add {
			action = "remove"
		}
		fmt.Fprintf(errOut, "usage: devclean config %s %s <id...>\n", which, action)
		return 2
	}

	cfg, usedPath, err := loadConfig(configPath)
	if err != nil {
		fmt.Fprintln(errOut, validateConfigPath(err, usedPath).Error())
		return 1
	}
	if usedPath == "" {
		fmt.Fprintln(errOut, "no config file found (use --config or run `devclean config init` first)")
		return 1
	}

	if which == "exclude" {
		if add {
			cfg.ExcludeIDs = addUniqueKeepOrder(cfg.ExcludeIDs, ids)
		} else {
			cfg.ExcludeIDs = removeAll(cfg.ExcludeIDs, ids)
		}
	} else {
		if add {
			cfg.IncludeIDs = addUniqueKeepOrder(cfg.IncludeIDs, ids)
		} else {
			cfg.IncludeIDs = removeAll(cfg.IncludeIDs, ids)
		}
	}

	if dryRun {
		fmt.Fprintln(out, "dry-run: pass without --dry-run to write changes")
		fmt.Fprintf(out, "config: %s\n", usedPath)
		return 0
	}

	if err := config.Save(usedPath, cfg); err != nil {
		fmt.Fprintln(errOut, err.Error())
		return 1
	}
	fmt.Fprintf(out, "updated config: %s\n", usedPath)
	return 0
}

func parseConfigEditArgs(args []string) (configPath string, dryRun bool, ids []string, err error) {
	// We intentionally allow flags to appear anywhere, e.g.
	// `devclean config exclude add foo --dry-run` or `... --dry-run foo`.
	onlyIDs := false
	for i := 0; i < len(args); i++ {
		a := strings.TrimSpace(args[i])
		if a == "" {
			continue
		}
		if a == "--" {
			onlyIDs = true
			continue
		}
		switch {
		case onlyIDs:
			// ID(s), allow comma-separated for convenience.
			for _, p := range strings.Split(a, ",") {
				p = strings.TrimSpace(p)
				if p != "" {
					ids = append(ids, p)
				}
			}
			continue
		case a == "--dry-run":
			dryRun = true
			continue
		case a == "--config":
			if i+1 >= len(args) {
				return "", false, nil, fmt.Errorf("--config requires a value")
			}
			i++
			configPath = strings.TrimSpace(args[i])
			if configPath == "" {
				return "", false, nil, fmt.Errorf("--config requires a non-empty value")
			}
			continue
		case strings.HasPrefix(a, "--config="):
			configPath = strings.TrimSpace(strings.TrimPrefix(a, "--config="))
			if configPath == "" {
				return "", false, nil, fmt.Errorf("--config requires a non-empty value")
			}
			continue
		case strings.HasPrefix(a, "-"):
			return "", false, nil, fmt.Errorf("unknown flag: %s", a)
		default:
			// ID(s), allow comma-separated for convenience.
			for _, p := range strings.Split(a, ",") {
				p = strings.TrimSpace(p)
				if p != "" {
					ids = append(ids, p)
				}
			}
		}
	}
	return configPath, dryRun, ids, nil
}

func addUniqueKeepOrder(existing []string, add []string) []string {
	seen := map[string]bool{}
	for _, e := range existing {
		seen[e] = true
	}
	out := append([]string{}, existing...)
	for _, a := range add {
		if !seen[a] {
			seen[a] = true
			out = append(out, a)
		}
	}
	return out
}

func removeAll(existing []string, remove []string) []string {
	rm := map[string]bool{}
	for _, r := range remove {
		rm[r] = true
	}
	var out []string
	for _, e := range existing {
		if !rm[e] {
			out = append(out, e)
		}
	}
	return out
}
