package auth_test

import (
	"os"
	"path/filepath"
	"testing"

	"encrypt-o-matic/internal/auth"
	"encrypt-o-matic/internal/config"

	"golang.org/x/crypto/bcrypt"
)

func TestVerifyStoredPassword(t *testing.T) {
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	hashPath, err := config.MasterHashPath()
	if err != nil {
		t.Fatal(err)
	}

	result, err := auth.VerifyStoredPassword("anything", hashPath)
	if result != auth.VerifyHashNotFound || err == nil {
		t.Fatalf("expected hash not found, got result=%v err=%v", result, err)
	}

	if err := os.MkdirAll(filepath.Dir(hashPath), 0o700); err != nil {
		t.Fatal(err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hashPath, hash, 0o600); err != nil {
		t.Fatal(err)
	}

	result, err = auth.VerifyStoredPassword("correct-password", hashPath)
	if result != auth.VerifyOK || err != nil {
		t.Fatalf("expected password OK, got result=%v err=%v", result, err)
	}

	result, err = auth.VerifyStoredPassword("wrong-password", hashPath)
	if result != auth.VerifyFailed || err != nil {
		t.Fatalf("expected verification failed, got result=%v err=%v", result, err)
	}
}
