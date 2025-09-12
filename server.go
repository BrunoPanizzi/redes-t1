package main

import (
	"bufio"
	"fmt"
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

func handleList(conn net.Conn) {
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

func handlePut(command *Command, conn net.Conn) {
	// payload = "<filename>\n<conteúdo>"
	parts := strings.SplitN(string(command.Payload), "\n", 2)
	if len(parts) < 2 {
		response := "PRBP PUT 5\nnot enough arguments"
		conn.Write([]byte(response))
		return
	}

	filename := strings.TrimSpace(parts[0])
	content := []byte(parts[1])
	filepath := storageDir + "/" + filename

	// Verificar se já existe
	if _, err := os.Stat(filepath); err == nil {
		response := "PRBP PUT 5\nfile already exists"
		conn.Write([]byte(response))
		return
	}

	// Criar e gravar no arquivo
	f, err := os.Create(filepath)
	if err != nil {
		response := "PRBP PUT 5\nerror creating file"
		conn.Write([]byte(response))
		return
	}
	defer f.Close()

	_, err = f.Write(content)
	if err != nil {
		response := "PRBP PUT 5\nerror writing to file"
		conn.Write([]byte(response))
		return
	}

	print("File saved: " + filepath + "\n")
	response := "PRBP PUT 2\nOK"
	conn.Write([]byte(response))
}

func handleQuit(conn net.Conn) {
	conn.Write([]byte("PRBP QUIT 0\n"))
	conn.Close()
}

func handleCommand(command *Command, conn net.Conn) {
	// fazer coisas aqui
	switch command.Method {
	case PUT:
		handlePut(command, conn)
	case LIST:
		handleList(conn)
	case QUIT:
		handleQuit(conn)
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

	// reads untill the first \n (the whole header)
	reader := bufio.NewReader(conn)
	header, err := reader.ReadBytes('\n')
	if err != nil {
		fmt.Printf("[%s] Could not read request header: %v\n", conn.RemoteAddr(), err)
		return
	}

	command, err := ParseHeader(string(header[:len(header)-1]), REQUEST)

	if err != nil {
		fmt.Printf("[%s] Could not parse header: %v\n", conn.RemoteAddr(), err)
		return
	}

	fmt.Printf("[%s] Command received: %v\n", conn.RemoteAddr(), command.String())
	fmt.Printf("[%s] Attempting to read %d bytes from payload\n", conn.RemoteAddr(), len(command.Payload))

	n, err := reader.Read(command.Payload)
	if err != nil {
		fmt.Printf("[%s] Could not read payload: %v\n", conn.RemoteAddr(), err)
		return
	}

	fmt.Printf("Read %d bytes from payload!\n%s\n", n, string(command.Payload[:n]))

	handleCommand(command, conn)

	conn.Write([]byte("Received your message!"))
}

func main() {
	fmt.Println("Starting the tcp server on the port 8080")
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}

	if err := os.MkdirAll(storageDir, 0755); err != nil {
		fmt.Println("Error creating storage directory:", err)
		return
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}
