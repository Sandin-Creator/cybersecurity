package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	DirName        = ".encryptomatic"
	MasterHashFile = "master.hash"
	MetadataSubdir = "metadata"
	BackupsSubdir  = "backups"
	EncryptedMagic = "EOMENC01"
)

// RootDir returns the path to .encryptomatic in the current working directory.
func RootDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to determine working directory: %w", err)
	}
	return filepath.Join(cwd, DirName), nil
}

// MasterHashPath returns the path to the bcrypt master password hash file.
func MasterHashPath() (string, error) {
	root, err := RootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, MasterHashFile), nil
}

// MetadataDir returns the path to the metadata storage directory.
func MetadataDir() (string, error) {
	root, err := RootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, MetadataSubdir), nil
}

// BackupsDir returns the path to the backup storage directory.
func BackupsDir() (string, error) {
	root, err := RootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, BackupsSubdir), nil
}
