package shared

import (
	"bufio"
	"net"
)

func SendLine(conn net.Conn, message string) error {
	_, err := conn.Write([]byte(message + "\n"))
	return err
}

func ReadLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return line[:len(line)-1], nil
}