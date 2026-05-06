package clean

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveAndValidatePathAllowedRoot_OutsideRootIsSkipped(t *testing.T) {
	root := t.TempDir()
	outside := "/tmp/outside-target"

	_, _, _, skip, reason := resolveAndValidatePathAllowedRoot(root, outside)
	if !skip {
		t.Fatalf("expected outside path to be skipped")
	}
	if reason != "outside allowed root directory" {
		t.Fatalf("unexpected skip reason: %s", reason)
	}
}

func TestResolveAndValidatePathAllowedRoot_GitPathIsProtected(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, ".git", "objects")

	_, _, _, skip, reason := resolveAndValidatePathAllowedRoot(root, p)
	if !skip {
		t.Fatalf("expected .git path to be skipped")
	}
	if reason != "protected .git path" {
		t.Fatalf("unexpected skip reason: %s", reason)
	}
}

func TestResolveAndValidatePathAllowedRoot_SymlinkToGitPathProtected(t *testing.T) {
	root := t.TempDir()
	gitPath := filepath.Join(root, ".git")
	if err := os.MkdirAll(gitPath, 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	link := filepath.Join(root, "link-to-git")
	if err := os.Symlink(gitPath, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	_, _, _, skip, reason := resolveAndValidatePathAllowedRoot(root, link)
	if !skip {
		t.Fatalf("expected symlink to .git path to be skipped")
	}
	if reason != "protected .git path (via symlink)" {
		t.Fatalf("unexpected skip reason: %s", reason)
	}
}

