package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/wangweicheng7/cleandev-cli/internal/cli"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	ctx := context.Background()

	// We keep root flags minimal; each subcommand has its own flagset.
	root := flag.NewFlagSet("cleandev", flag.ContinueOnError)
	root.SetOutput(os.Stderr)
	_ = root.Parse(os.Args[1:])

	args := root.Args()
	if len(args) == 0 {
		cli.PrintHelp(os.Stdout)
		return 2
	}

	cmd := args[0]
	cmdArgs := args[1:]

	switch cmd {
	case "scan":
		return cli.RunScan(ctx, cmdArgs, os.Stdout, os.Stderr)
	case "plan":
		// Alias for scan (kept for nicer semantics).
		return cli.RunScan(ctx, cmdArgs, os.Stdout, os.Stderr)
	case "clean":
		return cli.RunClean(ctx, cmdArgs, os.Stdout, os.Stderr)
	case "doctor":
		return cli.RunDoctor(ctx, cmdArgs, os.Stdout, os.Stderr)
	case "help", "-h", "--help":
		cli.PrintHelp(os.Stdout)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		cli.PrintHelp(os.Stderr)
		return 2
	}
}

