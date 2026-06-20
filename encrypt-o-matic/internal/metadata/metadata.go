package metadata

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"encrypt-o-matic/internal/config"
)

// FileMetadata holds everything required to restore an encrypted file exactly.
type FileMetadata struct {
	OriginalPath string    `json:"original_path"`
	Algorithm    string    `json:"algorithm"`
	Nonce        []byte    `json:"nonce"`
	Salt         []byte    `json:"salt"`
	PaddingSize  int64     `json:"padding_size"`
	OriginalHash string    `json:"original_sha256"`
	OriginalMode uint32    `json:"original_mode"`
	UnlockTime   time.Time `json:"unlock_time"`
	Compressed   bool      `json:"compressed"`
	EncryptedAt  time.Time `json:"encrypted_at"`
}

func metadataPathFor(targetPath string) (string, error) {
	abs, err := filepath.Abs(targetPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	sum := sha256.Sum256([]byte(abs))
	name := hex.EncodeToString(sum[:]) + ".json"

	metaDir, err := config.MetadataDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(metaDir, name), nil
}

// Save writes metadata JSON for an encrypted file.
func Save(meta *FileMetadata) error {
	path, err := metadataPathFor(meta.OriginalPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("file permission error: failed to save metadata: %w", err)
	}

	return nil
}

// Load reads metadata for a target path.
func Load(targetPath string) (*FileMetadata, error) {
	path, err := metadataPathFor(targetPath)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no encryption metadata found for %s", targetPath)
		}
		return nil, fmt.Errorf("file permission error: failed to read metadata: %w", err)
	}

	var meta FileMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("corrupted metadata: %w", err)
	}

	if meta.OriginalPath == "" || meta.Algorithm == "" || len(meta.Nonce) == 0 {
		return nil, fmt.Errorf("corrupted metadata: missing required fields")
	}

	return &meta, nil
}

// Remove deletes metadata for a target path.
func Remove(targetPath string) error {
	path, err := metadataPathFor(targetPath)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove metadata: %w", err)
	}
	return nil
}

// IsEncrypted reports whether metadata exists for the target path.
func IsEncrypted(targetPath string) bool {
	_, err := Load(targetPath)
	return err == nil
}
