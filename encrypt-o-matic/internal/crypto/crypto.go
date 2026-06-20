package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/twofish"
)

const (
	AlgoAES     = "AES"
	AlgoChaCha  = "ChaCha20"
	AlgoTwofish = "Twofish"

	pbkdf2Iterations = 100_000
	keySize          = 32
	saltSize         = 16
)

// NormalizeAlgorithm maps user input to a supported algorithm name.
func NormalizeAlgorithm(input string) (string, error) {
	switch input {
	case "AES", "aes":
		return AlgoAES, nil
	case "ChaCha20", "chacha20", "CHACHA20":
		return AlgoChaCha, nil
	case "Twofish", "twofish", "TWOFISH":
		return AlgoTwofish, nil
	default:
		return "", fmt.Errorf("unsupported algorithm: %s (supported: AES, ChaCha20, Twofish)", input)
	}
}

func deriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, pbkdf2Iterations, keySize, sha256.New)
}

func generateSalt() ([]byte, error) {
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

func generateNonce(size int) ([]byte, error) {
	nonce := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	return nonce, nil
}

// EncryptBytes encrypts plaintext with the selected algorithm using a password-derived key.
func EncryptBytes(algorithm, password string, plaintext []byte) (ciphertext, nonce, salt []byte, err error) {
	salt, err = generateSalt()
	if err != nil {
		return nil, nil, nil, err
	}

	key := deriveKey(password, salt)

	switch algorithm {
	case AlgoAES:
		ciphertext, nonce, err = encryptAES(key, plaintext)
	case AlgoChaCha:
		ciphertext, nonce, err = encryptChaCha(key, plaintext)
	case AlgoTwofish:
		ciphertext, nonce, err = encryptTwofish(key, plaintext)
	default:
		return nil, nil, nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
	if err != nil {
		return nil, nil, nil, err
	}
	return ciphertext, nonce, salt, nil
}

// DecryptBytes decrypts ciphertext with the selected algorithm using a password-derived key.
func DecryptBytes(algorithm, password string, ciphertext, nonce, salt []byte) ([]byte, error) {
	key := deriveKey(password, salt)

	switch algorithm {
	case AlgoAES:
		return decryptAES(key, nonce, ciphertext)
	case AlgoChaCha:
		return decryptChaCha(key, nonce, ciphertext)
	case AlgoTwofish:
		return decryptTwofish(key, nonce, ciphertext)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

func encryptAES(key, plaintext []byte) (ciphertext, nonce []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("AES setup failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("AES-GCM setup failed: %w", err)
	}

	nonce, err = generateNonce(gcm.NonceSize())
	if err != nil {
		return nil, nil, err
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

func decryptAES(key, nonce, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("AES setup failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("AES-GCM setup failed: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}
	return plaintext, nil
}

func encryptChaCha(key, plaintext []byte) (ciphertext, nonce []byte, err error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, nil, fmt.Errorf("ChaCha20 setup failed: %w", err)
	}

	nonce, err = generateNonce(aead.NonceSize())
	if err != nil {
		return nil, nil, err
	}

	ciphertext = aead.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

func decryptChaCha(key, nonce, ciphertext []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("ChaCha20 setup failed: %w", err)
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}
	return plaintext, nil
}

func encryptTwofish(key, plaintext []byte) (ciphertext, nonce []byte, err error) {
	block, err := twofish.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("Twofish setup failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("Twofish-GCM setup failed: %w", err)
	}

	nonce, err = generateNonce(gcm.NonceSize())
	if err != nil {
		return nil, nil, err
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

func decryptTwofish(key, nonce, ciphertext []byte) ([]byte, error) {
	block, err := twofish.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("Twofish setup failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("Twofish-GCM setup failed: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}
	return plaintext, nil
}
