package debug

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"encrypt-o-matic/internal/auth"
	"encrypt-o-matic/internal/config"
	"encrypt-o-matic/internal/metadata"
)

// RunVerifyPasswordCommand prompts for and verifies the master password.
func RunVerifyPasswordCommand() error {
	hashPath, err := config.MasterHashPath()
	if err != nil {
		return err
	}

	password, err := auth.PromptPassword("Enter master password: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	result, verifyErr := auth.VerifyStoredPassword(password, hashPath)
	switch result {
	case auth.VerifyOK:
		fmt.Println("Password OK")
		return nil
	case auth.VerifyHashNotFound:
		fmt.Println("Password INVALID")
		return verifyErr
	case auth.VerifyHashUnreadable:
		fmt.Println("Password INVALID")
		return verifyErr
	case auth.VerifyFailed:
		fmt.Println("Password INVALID")
		return errors.New("Password verification failed.")
	default:
		fmt.Println("Password INVALID")
		return verifyErr
	}
}

// RunDebugInfoCommand displays local configuration and storage state.
func RunDebugInfoCommand() error {
	configDir, err := config.RootDir()
	if err != nil {
		return err
	}

	hashPath, err := config.MasterHashPath()
	if err != nil {
		return err
	}

	fmt.Println("Encrypt-O-Matic Debug Information")
	fmt.Println()
	fmt.Println("Config directory:")
	fmt.Printf("  %s\n", configDir)
	fmt.Println()
	fmt.Println("Master hash:")
	if _, err := os.Stat(hashPath); os.IsNotExist(err) {
		fmt.Println("  Exists: No")
	} else if err != nil {
		fmt.Println("  Exists: Unknown (stat failed)")
	} else {
		fmt.Println("  Exists: Yes")
	}
	fmt.Printf("  Path: %s\n", hashPath)
	fmt.Println()

	metadataFiles, err := ListMetadataFiles()
	if err != nil {
		return err
	}
	fmt.Println("Metadata files:")
	if len(metadataFiles) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, entry := range metadataFiles {
			fmt.Printf("  %s\n", entry)
		}
	}
	fmt.Println()

	backupFiles, err := ListBackupFiles()
	if err != nil {
		return err
	}
	fmt.Println("Backup files:")
	if len(backupFiles) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, name := range backupFiles {
			fmt.Printf("  %s\n", name)
		}
	}

	return nil
}

// ListMetadataFiles returns metadata filenames with original paths when available.
func ListMetadataFiles() ([]string, error) {
	configDir, err := config.RootDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(configDir, config.MetadataSubdir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list metadata files: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		label := entry.Name()
		metaPath := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(metaPath)
		if err == nil {
			var meta metadata.FileMetadata
			if json.Unmarshal(data, &meta) == nil && meta.OriginalPath != "" {
				label = fmt.Sprintf("%s (%s)", entry.Name(), meta.OriginalPath)
			}
		}
		names = append(names, label)
	}

	sort.Strings(names)
	return names, nil
}

// ListBackupFiles returns backup filenames stored locally.
func ListBackupFiles() ([]string, error) {
	configDir, err := config.RootDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(configDir, config.BackupsSubdir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list backup files: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}

	sort.Strings(names)
	return names, nil
}
