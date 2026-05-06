package cli

import (
	"flag"
	"fmt"
	"strings"
)

type stringFlag struct {
	v   string
	set bool
}

func (f *stringFlag) String() string { return f.v }
func (f *stringFlag) Set(s string) error {
	f.v = s
	f.set = true
	return nil
}
func (f *stringFlag) IsSet() bool { return f.set }

type boolFlag struct {
	v   bool
	set bool
}

func (f *boolFlag) String() string { return fmt.Sprintf("%t", f.v) }
func (f *boolFlag) Set(s string) error {
	f.set = true
	s = strings.TrimSpace(strings.ToLower(s))
	// When user specifies `--flag` without a value, the flag package passes "true".
	if s == "true" {
		f.v = true
		return nil
	}
	if s == "false" {
		f.v = false
		return nil
	}
	return flag.ErrHelp
}
func (f *boolFlag) IsSet() bool { return f.set }
func (f *boolFlag) IsBoolFlag() bool { return true }

func newBoolFlag(def bool) *boolFlag { return &boolFlag{v: def} }

