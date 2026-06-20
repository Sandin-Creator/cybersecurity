package auth

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"encrypt-o-matic/internal/config"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

// VerifyResult indicates the outcome of a password verification attempt.
type VerifyResult int

const (
	VerifyOK VerifyResult = iota
	VerifyHashNotFound
	VerifyHashUnreadable
	VerifyFailed
)

// Authenticate prompts for the master password, creates or verifies the bcrypt
// hash stored in .encryptomatic/master.hash, and returns the password on success.
// This tool is for authorized educational testing only — never run it against
// paths you do not own or have explicit permission to modify.
func Authenticate() (string, error) {
	hashPath, err := config.MasterHashPath()
	if err != nil {
		return "", err
	}

	password, err := PromptPassword("Enter master password: ")
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	if _, err := os.Stat(hashPath); os.IsNotExist(err) {
		confirm, err := PromptPassword("Confirm master password: ")
		if err != nil {
			return "", fmt.Errorf("failed to read password confirmation: %w", err)
		}
		if password != confirm {
			return "", fmt.Errorf("passwords do not match")
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return "", fmt.Errorf("failed to create password hash: %w", err)
		}

		if err := os.MkdirAll(filepath.Dir(hashPath), 0o700); err != nil {
			return "", fmt.Errorf("failed to create auth directory: %w", err)
		}
		if err := os.WriteFile(hashPath, hash, 0o600); err != nil {
			return "", fmt.Errorf("failed to save password hash: %w", err)
		}

		fmt.Println("Master password created successfully.")
		return password, nil
	}

	result, verifyErr := VerifyStoredPassword(password, hashPath)
	switch result {
	case VerifyOK:
		fmt.Println("Password verified successfully.")
		return password, nil
	case VerifyHashNotFound:
		return "", verifyErr
	case VerifyHashUnreadable:
		return "", verifyErr
	case VerifyFailed:
		return "", errors.New("Password verification failed.")
	default:
		return "", verifyErr
	}
}

// VerifyStoredPassword checks a password against the bcrypt hash at hashPath.
func VerifyStoredPassword(password, hashPath string) (VerifyResult, error) {
	if _, err := os.Stat(hashPath); os.IsNotExist(err) {
		return VerifyHashNotFound, fmt.Errorf("master password hash file not found: %s", hashPath)
	} else if err != nil {
		return VerifyHashUnreadable, fmt.Errorf("master password hash file unreadable: %w", err)
	}

	stored, err := os.ReadFile(hashPath)
	if err != nil {
		return VerifyHashUnreadable, fmt.Errorf("master password hash file unreadable: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(stored, []byte(password)); err != nil {
		return VerifyFailed, nil
	}

	return VerifyOK, nil
}

// PromptPassword reads a password from stdin without echoing it.
func PromptPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	if term.IsTerminal(int(os.Stdin.Fd())) {
		bytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}

	var password string
	if _, err := fmt.Scanln(&password); err != nil {
		return "", err
	}
	return password, nil
}
