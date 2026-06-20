package fileops

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"encrypt-o-matic/internal/config"
	"encrypt-o-matic/internal/crypto"
	"encrypt-o-matic/internal/custom"
	"encrypt-o-matic/internal/metadata"
	"encrypt-o-matic/internal/timer"
)

// SHA256Hex returns the lowercase hex SHA-256 digest of data.
func SHA256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// CompressData gzip-compresses data.
func CompressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		return nil, fmt.Errorf("compression failed: %w", err)
	}
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("compression failed: %w", err)
	}
	return buf.Bytes(), nil
}

// DecompressData decompresses gzip data.
func DecompressData(data []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}
	defer gr.Close()

	out, err := io.ReadAll(gr)
	if err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}
	return out, nil
}

func readFileBytes(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("invalid path: file not found: %s", path)
		}
		return nil, fmt.Errorf("file permission error: failed to read %s: %w", path, err)
	}
	return data, nil
}

func createBackup(path string) (string, error) {
	data, err := readFileBytes(path)
	if err != nil {
		return "", err
	}

	backupRoot, err := config.BackupsDir()
	if err != nil {
		return "", err
	}

	base := filepath.Base(path)
	stamp := time.Now().UTC().Format("20060102-150405")
	backupPath := filepath.Join(backupRoot, fmt.Sprintf("%s.%s.bak", base, stamp))

	if err := os.MkdirAll(backupRoot, 0o700); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}
	if err := os.WriteFile(backupPath, data, 0o600); err != nil {
		return "", fmt.Errorf("file permission error: failed to create backup: %w", err)
	}

	return backupPath, nil
}

// GeneratePadding creates random padding of sizeMB megabytes.
func GeneratePadding(sizeMB int) ([]byte, error) {
	if sizeMB <= 0 {
		return nil, nil
	}

	size := int64(sizeMB) * 1024 * 1024
	padding := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, padding); err != nil {
		return nil, fmt.Errorf("failed to generate padding: %w", err)
	}
	return padding, nil
}

// BuildEncryptedPayload assembles the on-disk encrypted file format.
func BuildEncryptedPayload(ciphertext, padding []byte) []byte {
	magic := []byte(config.EncryptedMagic)
	payload := make([]byte, 0, len(magic)+len(ciphertext)+len(padding))
	payload = append(payload, magic...)
	payload = append(payload, ciphertext...)
	payload = append(payload, padding...)
	return payload
}

// StripEncryptedPayload removes the magic header and padding from encrypted data.
func StripEncryptedPayload(data []byte, paddingSize int64) ([]byte, error) {
	magic := []byte(config.EncryptedMagic)
	if len(data) < len(magic) {
		return nil, fmt.Errorf("corrupted encrypted file: missing header")
	}
	if string(data[:len(magic)]) != config.EncryptedMagic {
		return nil, fmt.Errorf("corrupted encrypted file: invalid header")
	}

	content := data[len(magic):]
	if paddingSize > 0 {
		if int64(len(content)) < paddingSize {
			return nil, fmt.Errorf("corrupted encrypted file: padding size mismatch")
		}
		content = content[:len(content)-int(paddingSize)]
	}

	if len(content) == 0 {
		return nil, fmt.Errorf("corrupted encrypted file: empty ciphertext")
	}

	return content, nil
}

func isSupportedTarget(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".exe")
}

func collectTargetFiles(targetPath string) ([]string, error) {
	info, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("invalid path: %s does not exist", targetPath)
		}
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	if !info.IsDir() {
		return []string{targetPath}, nil
	}

	var files []string
	err = filepath.WalkDir(targetPath, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("failed to access %s: %w", path, walkErr)
		}
		if d.IsDir() {
			return nil
		}
		if isSupportedTarget(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no supported files found in directory (only .exe files are encrypted by default)")
	}

	return files, nil
}

func encryptFile(path, algorithm, password string, paddingMB, durationMinutes int) error {
	if metadata.IsEncrypted(path) {
		meta, err := metadata.Load(path)
		if err != nil {
			return err
		}
		if timer.IsUnlockExpired(meta.UnlockTime) {
			fmt.Printf("%s is already encrypted but timer expired — decrypting automatically.\n", path)
			return decryptFile(path, password, true)
		}
		return fmt.Errorf("%s is already encrypted (%s)", path, timer.FormatUnlockStatus(meta.UnlockTime))
	}

	fmt.Printf("Processing: %s\n", path)

	original, err := readFileBytes(path)
	if err != nil {
		return err
	}
	originalInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	originalMode := originalInfo.Mode()
	originalHash := SHA256Hex(original)

	compressed, err := CompressData(original)
	if err != nil {
		return err
	}

	ciphertext, nonce, salt, err := crypto.EncryptBytes(algorithm, password, compressed)
	if err != nil {
		return err
	}

	padding, err := GeneratePadding(paddingMB)
	if err != nil {
		return err
	}

	payload := BuildEncryptedPayload(ciphertext, padding)

	backupPath, err := createBackup(path)
	if err != nil {
		return err
	}
	fmt.Printf("  Backup created: %s\n", backupPath)

	if err := os.WriteFile(path, payload, 0o600); err != nil {
		return fmt.Errorf("file permission error: failed to write encrypted file: %w", err)
	}

	meta := &metadata.FileMetadata{
		OriginalPath: path,
		Algorithm:    algorithm,
		Nonce:        nonce,
		Salt:         salt,
		PaddingSize:  int64(len(padding)),
		OriginalHash: originalHash,
		OriginalMode: uint32(originalMode),
		UnlockTime:   timer.ComputeUnlockTime(durationMinutes),
		Compressed:   true,
		EncryptedAt:  time.Now().UTC(),
	}

	if err := metadata.Save(meta); err != nil {
		return err
	}

	fmt.Printf("  Encrypted with %s (%s)\n", algorithm, timer.FormatUnlockStatus(meta.UnlockTime))
	return nil
}

func decryptFile(path, password string, automatic bool) error {
	meta, err := metadata.Load(path)
	if err != nil {
		return err
	}

	if !automatic {
		fmt.Printf("Decrypting: %s\n", path)
	} else {
		fmt.Printf("Auto-decrypting: %s\n", path)
	}

	encryptedFile, err := readFileBytes(path)
	if err != nil {
		return err
	}

	ciphertext, err := StripEncryptedPayload(encryptedFile, meta.PaddingSize)
	if err != nil {
		return err
	}

	compressed, err := crypto.DecryptBytes(meta.Algorithm, password, ciphertext, meta.Nonce, meta.Salt)
	if err != nil {
		return fmt.Errorf("decryption failed after successful authentication: %w", err)
	}

	var restored []byte
	if meta.Compressed {
		restored, err = DecompressData(compressed)
		if err != nil {
			return err
		}
	} else {
		restored = compressed
	}

	if SHA256Hex(restored) != meta.OriginalHash {
		return fmt.Errorf("hash mismatch after decryption: restored file does not match original SHA-256")
	}

	backupPath, err := createBackup(path)
	if err != nil {
		return err
	}
	fmt.Printf("  Encrypted copy backed up to: %s\n", backupPath)

	restoreMode := os.FileMode(meta.OriginalMode)
	if restoreMode == 0 {
		restoreMode = 0o600
	}

	if err := os.WriteFile(path, restored, restoreMode); err != nil {
		return fmt.Errorf("file permission error: failed to restore file: %w", err)
	}

	if err := metadata.Remove(path); err != nil {
		return err
	}

	fmt.Printf("  Restored successfully: %s\n", path)
	return nil
}

// EncryptTarget encrypts a file or directory of .exe files.
func EncryptTarget(targetPath, algorithm, password string, paddingMB, durationMinutes int, customRange string) error {
	if err := custom.RunOperation(customRange); err != nil {
		return err
	}

	files, err := collectTargetFiles(targetPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := encryptFile(file, algorithm, password, paddingMB, durationMinutes); err != nil {
			return err
		}
	}

	fmt.Println("Encryption complete.")
	return nil
}

// DecryptTarget decrypts a file or all encrypted files in a directory.
func DecryptTarget(targetPath, password string) error {
	info, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("invalid path: %s does not exist", targetPath)
		}
		return fmt.Errorf("invalid path: %w", err)
	}

	if info.IsDir() {
		var files []string
		walkErr := filepath.WalkDir(targetPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if metadata.IsEncrypted(path) {
				files = append(files, path)
			}
			return nil
		})
		if walkErr != nil {
			return walkErr
		}
		if len(files) == 0 {
			return fmt.Errorf("no encrypted files found in directory")
		}
		for _, file := range files {
			if err := decryptFile(file, password, false); err != nil {
				return err
			}
		}
		fmt.Println("Decryption complete.")
		return nil
	}

	if err := decryptFile(targetPath, password, false); err != nil {
		return err
	}
	fmt.Println("Decryption complete.")
	return nil
}

// CheckExpiredAutoDecrypt decrypts files whose timer has expired.
func CheckExpiredAutoDecrypt(targetPath, password string) error {
	files, err := collectTargetFiles(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var expired []string
	for _, file := range files {
		if !metadata.IsEncrypted(file) {
			continue
		}
		meta, err := metadata.Load(file)
		if err != nil {
			continue
		}
		if timer.IsUnlockExpired(meta.UnlockTime) {
			expired = append(expired, file)
		}
	}

	if len(expired) == 0 {
		return nil
	}

	fmt.Println("Expired timer detected — starting automatic decryption.")
	for _, file := range expired {
		if err := decryptFile(file, password, true); err != nil {
			return err
		}
	}
	return nil
}
