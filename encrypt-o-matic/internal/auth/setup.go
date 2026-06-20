package auth

import (
	"fmt"
	"os"
	"path/filepath"

	"encrypt-o-matic/internal/config"

	"golang.org/x/crypto/bcrypt"
)

// CreateMasterPassword stores a new bcrypt hash after confirmation.
func CreateMasterPassword(password, confirm string) error {
	if password != confirm {
		return fmt.Errorf("passwords do not match")
	}
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	hashPath, err := config.MasterHashPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(hashPath); err == nil {
		return fmt.Errorf("master password already configured")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to create password hash: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(hashPath), 0o700); err != nil {
		return fmt.Errorf("failed to create auth directory: %w", err)
	}
	if err := os.WriteFile(hashPath, hash, 0o600); err != nil {
		return fmt.Errorf("failed to save password hash: %w", err)
	}
	return nil
}

// MasterPasswordConfigured reports whether master.hash exists.
func MasterPasswordConfigured() (bool, error) {
	hashPath, err := config.MasterHashPath()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(hashPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}
