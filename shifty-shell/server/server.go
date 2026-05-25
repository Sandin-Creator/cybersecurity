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
	listener, err := net.Listen("tcp", ":4444")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server listening on port 4444")

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Client connected:", conn.RemoteAddr())

	reader := bufio.NewReader(conn)

	serverPrivate, serverPublic, err := secure.GenerateKeyPair()
	if err != nil {
		panic(err)
	}

	err = shared.SendLine(conn, hex.EncodeToString(serverPublic))
	if err != nil {
		panic(err)
	}

	clientPublicHex, err := shared.ReadLine(reader)
	if err != nil {
		panic(err)
	}

	clientPublic, err := hex.DecodeString(clientPublicHex)
	if err != nil {
		panic(err)
	}

	sharedSecret, err := secure.GenerateSharedSecret(serverPrivate, clientPublic)
	if err != nil {
		panic(err)
	}

	sessionKey, err := secure.DeriveAESKey(sharedSecret)
	if err != nil {
		panic(err)
	}

	fmt.Println("Session key fingerprint:", hex.EncodeToString(sessionKey[:8]))

	message := "hello from server"

	encryptedMessage, err := secure.EncryptMessage(sessionKey, message)
	if err != nil {
		panic(err)
	}

	err = shared.SendLine(conn, encryptedMessage)
	if err != nil {
		panic(err)
	}

	fmt.Println("Encrypted message sent:", encryptedMessage)

	encryptedReply, err := shared.ReadLine(reader)
	if err != nil {
		panic(err)
	}

	fmt.Println("Encrypted reply received:", encryptedReply)

	decryptedReply, err := secure.DecryptMessage(sessionKey, encryptedReply)
	if err != nil {
		panic(err)
	}

	fmt.Println("Decrypted reply:", decryptedReply)
}