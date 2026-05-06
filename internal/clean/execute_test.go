package clean

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestExecute_DryRunDoesNotDelete(t *testing.T) {
	home, err := HomeDir()
	if err != nil {
		t.Fatalf("home dir: %v", err)
	}

	cacheDir := filepath.Join(home, "Library", "Caches", "go-build")
	testDir := filepath.Join(cacheDir, "devclean-test-dryrun")
	t.Cleanup(func() {
		_ = os.RemoveAll(testDir)
	})
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		t.Fatalf("mkdir cache dir: %v", err)
	}
	f := filepath.Join(testDir, "artifact.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	res, err := Execute(context.Background(), ExecuteOptions{
		DryRun:   true,
		Profile:  ProfileSafe,
		WithSize: false,
	})
	if err != nil {
		t.Fatalf("execute dry-run: %v", err)
	}
	if len(res.DeletedIDs) == 0 {
		t.Fatalf("expected dry-run to report at least one deletion candidate")
	}
	if _, err := os.Stat(f); err != nil {
		t.Fatalf("expected file to remain after dry-run, stat err: %v", err)
	}
}

func TestExecute_RequiresConfirmOrDryRun(t *testing.T) {
	_, err := Execute(context.Background(), ExecuteOptions{
		Profile:  ProfileSafe,
		WithSize: false,
	})
	if err == nil {
		t.Fatalf("expected error when neither dry-run nor confirm is provided")
	}
}

