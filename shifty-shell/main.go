package main

import (
	"encoding/hex"
	"fmt"
	"shifty-shell/secure"
)

func main() {
	serverPrivate, serverPublic, err := secure.GenerateKeyPair()
	if err != nil {
		panic(err)
	}

	clientPrivate, clientPublic, err := secure.GenerateKeyPair()
	if err != nil {
		panic(err)
	}

	serverSecret, err := secure.GenerateSharedSecret(serverPrivate, clientPublic)
	if err != nil {
		panic(err)
	}

	clientSecret, err := secure.GenerateSharedSecret(clientPrivate, serverPublic)
	if err != nil {
		panic(err)
	}

	serverAESKey, err := secure.DeriveAESKey(serverSecret)
	if err != nil {
		panic(err)
	}

	clientAESKey, err := secure.DeriveAESKey(clientSecret)
	if err != nil {
		panic(err)
	}

	fmt.Println("Server AES key fingerprint:", hex.EncodeToString(serverAESKey[:8]))
	fmt.Println("Client AES key fingerprint:", hex.EncodeToString(clientAESKey[:8]))
}