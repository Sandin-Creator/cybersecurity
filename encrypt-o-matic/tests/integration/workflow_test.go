package integration_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"encrypt-o-matic/internal/crypto"
	"encrypt-o-matic/internal/fileops"
)

func loadTestBinary(t *testing.T) []byte {
	t.Helper()

	path := filepath.Join("..", "testdata", "demo.exe")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}
	return data
}

func TestFileWorkflowAESChaChaTwofish(t *testing.T) {
	minimalPE := loadTestBinary(t)

	dir := t.TempDir()
	testFile := filepath.Join(dir, "demo.exe")
	if err := os.WriteFile(testFile, minimalPE, 0o755); err != nil {
		t.Fatalf("create test file: %v", err)
	}

	originalHash := fileops.SHA256Hex(minimalPE)
	password := "workflow-test-password"

	algorithms := []string{crypto.AlgoAES, crypto.AlgoChaCha, crypto.AlgoTwofish}
	for _, algorithm := range algorithms {
		t.Run(algorithm, func(t *testing.T) {
			current, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("read test file: %v", err)
			}
			if fileops.SHA256Hex(current) != originalHash {
				t.Fatalf("test file was not restored before %s run", algorithm)
			}

			compressed, err := fileops.CompressData(current)
			if err != nil {
				t.Fatalf("compress: %v", err)
			}

			ciphertext, nonce, salt, err := crypto.EncryptBytes(algorithm, password, compressed)
			if err != nil {
				t.Fatalf("encrypt: %v", err)
			}

			padding, err := fileops.GeneratePadding(0)
			if err != nil {
				t.Fatalf("padding: %v", err)
			}

			payload := fileops.BuildEncryptedPayload(ciphertext, padding)
			if err := os.WriteFile(testFile, payload, 0o755); err != nil {
				t.Fatalf("write encrypted file: %v", err)
			}

			encryptedOnDisk, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("read encrypted file: %v", err)
			}

			stripped, err := fileops.StripEncryptedPayload(encryptedOnDisk, int64(len(padding)))
			if err != nil {
				t.Fatalf("strip payload: %v", err)
			}

			decryptedCompressed, err := crypto.DecryptBytes(algorithm, password, stripped, nonce, salt)
			if err != nil {
				t.Fatalf("decrypt: %v", err)
			}

			restored, err := fileops.DecompressData(decryptedCompressed)
			if err != nil {
				t.Fatalf("decompress: %v", err)
			}

			if fileops.SHA256Hex(restored) != originalHash {
				t.Fatal("hash mismatch after decryption")
			}

			if !bytes.Equal(restored, minimalPE) {
				t.Fatal("restored bytes do not match original test binary")
			}

			if err := os.WriteFile(testFile, restored, 0o755); err != nil {
				t.Fatalf("restore file: %v", err)
			}
		})
	}
}
