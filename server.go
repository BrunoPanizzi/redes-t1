package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

// <protocol> <metodo> [payload_size]\n
// [payload]

// cliente  -> PRBP LIST\n
// servidor -> PRBP LIST <payload_size>\n
//             [payload]

// cliente  -> PRBP PUT <payload_size>\n
//             [payload]
// servidor -> PRBP PUT <payload_size>\n
//             <request_status>

// cliente  -> PRBP QUIT\n
// servidor -> PRBP QUIT\n

type Method int

const (
	LIST Method = iota
	PUT
	QUIT
)

type CommandType int

const (
	REQUEST CommandType = iota
	RESPONSE
)

const storageDir = "storage/"

type Command struct {
	Type        CommandType
	Method      Method
	PayloadSize int
	Payload     []byte
}

func serverHandleList(conn net.Conn) {
	files, err := os.ReadDir(storageDir)
	if err != nil {
		fmt.Printf("Error reading storage directory: %v\n", err)
		conn.Write([]byte("PRBP LIST 0\n"))
		return
	}

	var builder strings.Builder
	for _, file := range files {
		if !file.IsDir() {
			builder.WriteString(file.Name() + "\n")
		}
	}

	payload := builder.String()
	response := fmt.Sprintf("PRBP LIST %d\n%s", len(payload), payload)
	conn.Write([]byte(response))
}

func serverHandlePut(command *Command, conn net.Conn) {
	if len(command.Payload) == 0 {
		response := "PRBP PUT 19\nError: No payload"
		conn.Write([]byte(response))
		return
	}

	parts := strings.SplitN(string(command.Payload), "\n", 2)
	if len(parts) < 2 {
		response := "PRBP PUT 26\nError: Invalid payload format"
		conn.Write([]byte(response))
		return
	}

	filename := strings.TrimSpace(parts[0])
	content := []byte(parts[1])
	filepath := storageDir + filename

	f, err := os.Create(filepath) // cria ou sobrescreve
	if err != nil {
		response := "PRBP PUT 24\nError: Creating file"
		conn.Write([]byte(response))
		return
	}
	defer f.Close()

	_, err = f.Write(content)
	if err != nil {
		response := "PRBP PUT 24\nError: Writing file"
		conn.Write([]byte(response))
		return
	}

	fmt.Printf("File saved (overwritten if existed): %s (%d bytes)\n", filepath, len(content))
	response := "PRBP PUT 2\nOK"
	conn.Write([]byte(response))
}

func serverHandleQuit(conn net.Conn) {
	response := "PRBP QUIT 0\n"
	conn.Write([]byte(response))
	conn.Close()
	fmt.Printf("[%s] Client disconnected via QUIT\n", conn.RemoteAddr())
}

func ParseMethod(s string) (Method, error) {
	switch s {
	case "LIST":
		return LIST, nil
	case "PUT":
		return PUT, nil
	case "QUIT":
		return QUIT, nil
	default:
		return -1, fmt.Errorf("invalid method: %s", s)
	}
}

func ParseHeader(data string, commandType CommandType) (*Command, error) {
	parts := strings.SplitN(data, " ", 4)

	if parts[0] != "PRBP" {
		return nil, fmt.Errorf("invalid protocol")
	}

	method, err := ParseMethod(parts[1])

	if err != nil {
		return nil, err
	}

	var payloadSize int

	if len(parts) > 2 {
		payloadSize, err = strconv.Atoi(parts[2])
		if err != nil {
			return nil, err
		}
	}

	cmd := &Command{
		Type:        commandType,
		Method:      method,
		PayloadSize: payloadSize,
		Payload:     make([]byte, payloadSize),
	}

	return cmd, nil
}

func (c *Command) String() string {
	return fmt.Sprintf("Type: %v, Method: %v, PayloadSize: %d, Payload: %s{}",
		c.Type, c.Method, c.PayloadSize, string(c.Payload))
}

func handleCommand(command *Command, conn net.Conn) {
	switch command.Method {
	case PUT:
		serverHandlePut(command, conn)
	case LIST:
		serverHandleList(conn)
	case QUIT:
		serverHandleQuit(conn)
	default:
		fmt.Printf("Unknown method: %v\n", command.Method)
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		fmt.Printf("[%s] Connection closed\n", conn.RemoteAddr())
	}()
	fmt.Printf("[%s] Client connected\n", conn.RemoteAddr())

	reader := bufio.NewReader(conn)

	for {
		header, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("[%s] Error reading header: %v\n", conn.RemoteAddr(), err)
			}
			return
		}

		command, err := ParseHeader(string(header[:len(header)-1]), REQUEST)
		if err != nil {
			fmt.Printf("[%s] Could not parse header: %v\n", conn.RemoteAddr(), err)
			return
		}

		if command.PayloadSize > 0 {
			_, err := io.ReadFull(reader, command.Payload)
			if err != nil {
				fmt.Printf("[%s] Could not read payload: %v\n", conn.RemoteAddr(), err)
				return
			}
		}

		fmt.Printf("[%s] Command received: %v\n", conn.RemoteAddr(), command.String())
		handleCommand(command, conn)

		if command.Method == QUIT {
			return
		}
	}
}

func main() {
	fmt.Println("Starting TCP server on port 8080")

	if err := os.MkdirAll(storageDir, 0755); err != nil {
		fmt.Println("Error creating storage directory:", err)
		return
	}

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()

	fmt.Println("Server listening on port 8080")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}
