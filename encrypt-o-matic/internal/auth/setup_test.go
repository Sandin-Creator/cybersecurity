package auth

import (
	"os"
	"path/filepath"
	"testing"

	"encrypt-o-matic/internal/config"
)

func TestCreateMasterPassword(t *testing.T) {
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	ok, err := MasterPasswordConfigured()
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected no master password on fresh config")
	}

	if err := CreateMasterPassword("secret123", "secret123"); err != nil {
		t.Fatalf("CreateMasterPassword: %v", err)
	}

	ok, err = MasterPasswordConfigured()
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected master password to be configured")
	}

	if err := CreateMasterPassword("other", "other"); err == nil {
		t.Fatal("expected error when password already configured")
	}

	hashPath, err := config.MasterHashPath()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(hashPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("hash file is empty")
	}
	if filepath.Base(hashPath) != "master.hash" {
		t.Fatalf("unexpected hash filename: %s", hashPath)
	}
}

func TestCreateMasterPasswordMismatch(t *testing.T) {
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	if err := CreateMasterPassword("a", "b"); err == nil {
		t.Fatal("expected mismatch error")
	}
}
