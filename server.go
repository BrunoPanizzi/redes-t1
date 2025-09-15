package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/BrunoPanizzi/redes_t1/prbp"
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

const storageDir = "storage/"

func serverHandleList(request *prbp.Command) (response *prbp.Command) {
	var message string

	// this will create the response and set the payload to the message variable at the end of the function
	defer func() {
		response = prbp.NewCommand(prbp.RESPONSE, prbp.LIST).SetPayload([]byte(message))
	}()

	files, err := os.ReadDir(storageDir)
	if err != nil {
		fmt.Printf("Error reading storage directory: %v\n", err)
		message = "Error: Could not read storage directory"
		return
	}

	var builder strings.Builder
	for _, file := range files {
		if !file.IsDir() {
			builder.WriteString(file.Name() + "\n")
		}
	}

	message = builder.String()
	return
}

func serverHandlePut(request *prbp.Command) (response *prbp.Command) {
	var message string

	// this will create the response and set the payload to the message variable at the end of the function
	defer func() {
		response = prbp.NewCommand(prbp.RESPONSE, prbp.PUT).SetPayload([]byte(message))
	}()

	if len(request.Payload) == 0 {
		message = "Error: No payload"
		return
	}

	parts := strings.SplitN(string(request.Payload), "\n", 2)
	if len(parts) < 2 {
		message = "Error: Payload should be <filename>\\n<file_content>"
		return
	}

	filename := strings.TrimSpace(parts[0])
	content := []byte(parts[1])
	filepath := storageDir + filename

	f, err := os.Create(filepath) // cria ou sobrescreve
	if err != nil {
		message = "Error: Could not create the file"
		return
	}
	defer f.Close()

	_, err = f.Write(content)
	if err != nil {
		message = "Error: Could not write the file"
		return
	}

	fmt.Printf("File saved (overwritten if existed): %s (%d bytes)\n", filepath, len(content))

	message = "OK"
	return
}

func serverHandleQuit(request *prbp.Command) (response *prbp.Command) {
	response = prbp.NewCommand(prbp.RESPONSE, prbp.QUIT).SetPayload([]byte("OK"))
	return
}

func handleCommand(request *prbp.Command) *prbp.Command {
	switch request.Method {
	case prbp.PUT:
		return serverHandlePut(request)
	case prbp.LIST:
		return serverHandleList(request)
	case prbp.QUIT:
		return serverHandleQuit(request)
	default:
		fmt.Printf("Unknown method: %v\n", request.Method)
		return nil
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		fmt.Printf("[%s] Connection closed\n", conn.RemoteAddr())
	}()
	fmt.Printf("[%s] Client connected\n", conn.RemoteAddr())

	for {
		request, err := prbp.ParseCommand(conn, prbp.REQUEST)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("[%s] Error reading header: %v\n", conn.RemoteAddr(), err)
			}
			return
		}

		fmt.Printf("[%s] Command received: %v\n", conn.RemoteAddr(), request.String())
		response := handleCommand(request)

		if response != nil {
			n, err := conn.Write(response.Bytes())
			if err != nil {
				fmt.Printf("[%s] Error sending response: %v\n", conn.RemoteAddr(), err)
				return
			}
			fmt.Printf("[%s] Sent response (%d bytes): %v\n", conn.RemoteAddr(), n, response.String())
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
