package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Metrics struct {
	StartTime     time.Time
	EndTime       time.Time
	BytesSent     int
	BytesReceived int
}

const storage = "client_storage/"

func parseServerResponse(header string) (method string, payloadSize int, err error) {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 3)
	if len(parts) < 2 || parts[0] != "PRBP" {
		return "", 0, fmt.Errorf("invalid protocol header")
	}

	method = parts[1]

	if len(parts) > 2 {
		payloadSize, err = strconv.Atoi(parts[2])
		if err != nil {
			return "", 0, fmt.Errorf("invalid payload size: %v", err)
		}
	}

	return method, payloadSize, nil
}

func handleList(conn net.Conn, metrics *Metrics) {
	msg := "PRBP LIST\n"
	n, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("Error sending LIST command:", err)
		return
	}
	metrics.BytesSent += n

	reader := bufio.NewReader(conn)
	header, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}
	metrics.BytesReceived += len(header)

	method, payloadSize, err := parseServerResponse(header)
	if err != nil {
		fmt.Println("Error parsing response:", err)
		return
	}

	if method != "LIST" {
		fmt.Printf("Unexpected response method: %s (expected LIST)\n", method)
		return
	}

	if payloadSize > 0 {
		payload := make([]byte, payloadSize)
		n, err := io.ReadFull(reader, payload)
		if err != nil {
			fmt.Println("Error reading response payload:", err)
			return
		}
		metrics.BytesReceived += n
		fmt.Printf("Files on server:\n%s", string(payload))
	} else {
		fmt.Println("No files found on server.")
	}
}

func handlePut(conn net.Conn, filename string, metrics *Metrics) {
	completePath := filepath.Join(storage, filename)
	content, err := os.ReadFile(completePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	payload := fmt.Sprintf("%s\n%s", filepath.Base(filename), string(content))
	header := fmt.Sprintf("PRBP PUT %d\n", len(payload))

	completeMessage := header + payload
	n, err := conn.Write([]byte(completeMessage))
	if err != nil {
		fmt.Println("Error sending PUT command:", err)
		return
	}
	metrics.BytesSent += n

	reader := bufio.NewReader(conn)
	responseHeader, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}
	metrics.BytesReceived += len(responseHeader)

	method, payloadSize, err := parseServerResponse(responseHeader)
	if err != nil {
		fmt.Println("Error parsing response:", err)
		return
	}

	if method != "PUT" {
		fmt.Printf("Unexpected response method: %s (expected PUT)\n", method)
		return
	}

	if payloadSize > 0 {
		responsePayload := make([]byte, payloadSize)
		n, err := io.ReadFull(reader, responsePayload)
		if err != nil {
			fmt.Println("Error reading response payload:", err)
			return
		}
		metrics.BytesReceived += n
		fmt.Printf("Server response: %s\n", string(responsePayload))
	}
}

func handleQuit(conn net.Conn, metrics *Metrics) {
	msg := "PRBP QUIT\n"
	n, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("Error sending QUIT command:", err)
		return
	}
	metrics.BytesSent += n

	// Ler resposta do servidor (opcional)
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err == nil {
		fmt.Println("Server response:", strings.TrimSpace(response))
		metrics.BytesReceived += len(response)
	}

	conn.Close()
	fmt.Println("Disconnected from server.")
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: client <host> <port>")
		fmt.Println("Example: client localhost 8080")
		return
	}

	if err := os.MkdirAll(storage, 0755); err != nil {
		fmt.Println("Error creating client storage directory:", err)
		return
	}

	host := os.Args[1]
	port := os.Args[2]
	address := fmt.Sprintf("%s:%s", host, port)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	metrics := Metrics{StartTime: time.Now()}
	defer func() {
		metrics.EndTime = time.Now()
		duration := metrics.EndTime.Sub(metrics.StartTime).Seconds()
		fmt.Printf("\nSession Metrics:\n")
		fmt.Printf("Duration: %.2f seconds\n", duration)
		fmt.Printf("Bytes Sent: %d bytes\n", metrics.BytesSent)
		fmt.Printf("Bytes Received: %d bytes\n", metrics.BytesReceived)
		if duration > 0 {
			fmt.Printf("Throughput: %.2f bytes/sec\n", float64(metrics.BytesSent)/duration)
		}
	}()

	fmt.Printf("Connected to server at %s\n", address)
	fmt.Println("Available commands: LIST, PUT <filename>, QUIT")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		command := strings.ToUpper(parts[0])
		switch command {
		case "LIST":
			handleList(conn, &metrics)
		case "PUT":
			if len(parts) < 2 {
				fmt.Println("Usage: PUT <filename>")
				continue
			}
			handlePut(conn, parts[1], &metrics)
		case "QUIT":
			handleQuit(conn, &metrics)
			return
		default:
			fmt.Println("Unknown command. Available commands: LIST, PUT <filename>, QUIT")
		}
	}
}
