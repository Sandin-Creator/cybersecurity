package secure

import (
	"crypto/sha256"
	"io"
	"crypto/rand"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/curve25519"
)

func GenerateKeyPair() (privateKey []byte, publicKey []byte, err error) {
	privateKey = make([]byte, 32)

	_, err = rand.Read(privateKey)
	if err != nil {
		return nil, nil, err
	}

	publicKey, err = curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, publicKey, nil
}

func GenerateSharedSecret(privateKey []byte, peerPublicKey []byte) ([]byte, error) {
	return curve25519.X25519(privateKey, peerPublicKey)
}

func DeriveAESKey(sharedSecret []byte) ([]byte, error) {
	hkdfReader := hkdf.New(
		sha256.New,
		sharedSecret,
		nil,
		[]byte("shifty-shell-session-key"),
	)

	key := make([]byte, 32)

	_, err := io.ReadFull(hkdfReader, key)
	if err != nil {
		return nil, err
	}

	return key, nil
}