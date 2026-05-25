package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"net"
	"shifty-shell/secure"
	"shifty-shell/shared"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:4444")
	if err != nil {
		fmt.Println("Connection failed:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to server")

	reader := bufio.NewReader(conn)

	clientPrivate, clientPublic, err := secure.GenerateKeyPair()
	if err != nil {
		panic(err)
	}

	serverPublicHex, err := shared.ReadLine(reader)
	if err != nil {
		panic(err)
	}

	serverPublic, err := hex.DecodeString(serverPublicHex)
	if err != nil {
		panic(err)
	}

	err = shared.SendLine(conn, hex.EncodeToString(clientPublic))
	if err != nil {
		panic(err)
	}

	sharedSecret, err := secure.GenerateSharedSecret(clientPrivate, serverPublic)
	if err != nil {
		panic(err)
	}

	sessionKey, err := secure.DeriveAESKey(sharedSecret)
	if err != nil {
		panic(err)
	}

	fmt.Println("Session key fingerprint:", hex.EncodeToString(sessionKey[:8]))

	encryptedMessage, err := shared.ReadLine(reader)
	if err != nil {
		panic(err)
	}

	fmt.Println("Encrypted received:", encryptedMessage)

	decryptedMessage, err := secure.DecryptMessage(sessionKey, encryptedMessage)
	if err != nil {
		panic(err)
	}

	fmt.Println("Decrypted message:", decryptedMessage)

	reply := "hello back from client"

	encryptedReply, err := secure.EncryptMessage(sessionKey, reply)
	if err != nil {
		panic(err)
	}

	err = shared.SendLine(conn, encryptedReply)
	if err != nil {
		panic(err)
	}

	fmt.Println("Encrypted reply sent:", encryptedReply)
}