package web

import (
	"os"
	"path/filepath"
	"testing"
)

func chdirForTest(t *testing.T, dir string) {
	t.Helper()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
}

func TestBrowseDirectoryWorkingDir(t *testing.T) {
	dir := t.TempDir()
	exePath := filepath.Join(dir, "sample.exe")
	if err := os.WriteFile(exePath, []byte("MZ"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "nested")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	chdirForTest(t, dir)

	resp, err := BrowseDirectory(dir)
	if err != nil {
		t.Fatalf("BrowseDirectory: %v", err)
	}
	if resp.CurrentPath == "" {
		t.Fatal("expected current path")
	}
	if len(resp.Folders) != 1 || resp.Folders[0].Name != "nested" {
		t.Fatalf("expected nested folder, got %+v", resp.Folders)
	}
	if len(resp.Files) != 1 || resp.Files[0].Name != "sample.exe" {
		t.Fatalf("expected sample.exe, got %+v", resp.Files)
	}
	if len(resp.Roots) == 0 {
		t.Fatal("expected roots in response")
	}
}

func TestBrowseDirectoryRejectsOutsideRoots(t *testing.T) {
	dir := t.TempDir()
	chdirForTest(t, dir)

	_, err := BrowseDirectory("/etc")
	if err == nil {
		t.Fatal("expected error for path outside allowed roots")
	}
}

func TestBrowseDirectoryParentPath(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "child")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	chdirForTest(t, dir)

	resp, err := BrowseDirectory(sub)
	if err != nil {
		t.Fatal(err)
	}
	if resp.ParentPath == "" {
		t.Fatal("expected parent path")
	}
}
