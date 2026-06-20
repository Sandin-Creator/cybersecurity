package crypto_test

import (
	"bytes"
	"testing"

	"encrypt-o-matic/internal/crypto"
	"encrypt-o-matic/internal/fileops"
)

// minimalPE is a tiny valid-ish PE stub used as a test binary payload.
var minimalPE = []byte{
	'M', 'Z', 0x90, 0x00, 0x03, 0x00, 0x00, 0x00,
	0x04, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0x00, 0x00,
	0xB8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x00,
	0x0E, 0x1F, 0xBA, 0x0E, 0x00, 0xB4, 0x09, 0xCD,
	0x21, 0xB8, 0x01, 0x4C, 0xCD, 0x21, 'T', 'e',
	's', 't', 'E', 'X', 'E', '-', 'e', 'n', 'c',
	'r', 'y', 'p', 't', '-', 'o', '-', 'm', 'a', 't',
	'i', 'c', 0x00,
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	password := "test-master-password"
	originalHash := fileops.SHA256Hex(minimalPE)

	algorithms := []string{crypto.AlgoAES, crypto.AlgoChaCha, crypto.AlgoTwofish}
	for _, algorithm := range algorithms {
		t.Run(algorithm, func(t *testing.T) {
			compressed, err := fileops.CompressData(minimalPE)
			if err != nil {
				t.Fatalf("compress: %v", err)
			}

			ciphertext, nonce, salt, err := crypto.EncryptBytes(algorithm, password, compressed)
			if err != nil {
				t.Fatalf("encrypt: %v", err)
			}

			decryptedCompressed, err := crypto.DecryptBytes(algorithm, password, ciphertext, nonce, salt)
			if err != nil {
				t.Fatalf("decrypt: %v", err)
			}

			restored, err := fileops.DecompressData(decryptedCompressed)
			if err != nil {
				t.Fatalf("decompress: %v", err)
			}

			if !bytes.Equal(restored, minimalPE) {
				t.Fatal("restored bytes do not match original test binary")
			}

			if fileops.SHA256Hex(restored) != originalHash {
				t.Fatal("hash mismatch after decryption")
			}
		})
	}
}

func TestTwofishWrongPasswordFails(t *testing.T) {
	ciphertext, nonce, salt, err := crypto.EncryptBytes(crypto.AlgoTwofish, "correct-password", []byte("payload"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	if _, err := crypto.DecryptBytes(crypto.AlgoTwofish, "wrong-password", ciphertext, nonce, salt); err == nil {
		t.Fatal("expected decryption to fail with wrong password")
	}
}
