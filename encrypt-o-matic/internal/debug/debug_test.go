package debug_test

import (
	"os"
	"path/filepath"
	"testing"

	"encrypt-o-matic/internal/config"
	"encrypt-o-matic/internal/debug"
)

func TestDebugInfoListsMetadataAndBackups(t *testing.T) {
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	configDir, err := config.RootDir()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(configDir, config.MetadataSubdir), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(configDir, config.BackupsSubdir), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, config.MetadataSubdir, "abc123.json"), []byte(`{"original_path":"C:\\demo.exe","algorithm":"AES","nonce":"AA=="}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, config.BackupsSubdir, "demo.exe.20260620-190622.bak"), []byte("backup"), 0o600); err != nil {
		t.Fatal(err)
	}

	metadataFiles, err := debug.ListMetadataFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(metadataFiles) != 1 {
		t.Fatalf("expected 1 metadata entry, got %d", len(metadataFiles))
	}

	backupFiles, err := debug.ListBackupFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(backupFiles) != 1 || backupFiles[0] != "demo.exe.20260620-190622.bak" {
		t.Fatalf("unexpected backup list: %#v", backupFiles)
	}
}
