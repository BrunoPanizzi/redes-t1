package main

import (
	"fmt"
	"net"
)

func talk(conn net.Conn) {
	defer func() {
		conn.Close()
		fmt.Printf("[%s] Connection closed\n", conn.RemoteAddr())
	}()

	fmt.Printf("[%s] Connected to server\n", conn.RemoteAddr())

	// now the protocol have this format: PRBP <METHOD> <PAYLOAD_SIZE>\n<FILENAME>\n<PAYLOAD>
	// Example: PRBP PUT 23\nhello.txt\nHello world!!!
	// Send a PUT command to the server
	// Here we are sending a file named "hello.txt" with the content "Hello world!!!"
	// I've implemented this way cause we needed a division to know which name the file will have.
	// The number of bytes is 23 because "hello.txt\nHello world!!!" has 23 bytes
	written, err := conn.Write([]byte("PRBP PUT 23\nhello.txt\nHello world!!!"))

	if err != nil {
		fmt.Printf("[%s] Error writing to server: %v\n", conn.RemoteAddr(), err)
		return
	}

	fmt.Printf("[%s] Wrote %d bytes to server\n", conn.RemoteAddr(), written)

	// Read response from server
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("[%s] Error reading from server: %v\n", conn.RemoteAddr(), err)
		return
	}
	fmt.Printf("[%s] Received from server: %s\n", conn.RemoteAddr(), string(buffer[:n]))
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}

	talk(conn)
}
