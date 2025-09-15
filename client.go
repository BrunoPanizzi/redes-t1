package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BrunoPanizzi/redes_t1/prbp"
)

type Metrics struct {
	StartTime        time.Time
	EndTime          time.Time
	TransmittingTime time.Duration
	BytesSent        int
	BytesReceived    int
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

func timeIt(f func()) time.Duration {
	start := time.Now()
	f()
	return time.Since(start)
}

// Human-readable byte formatting helpers
func humanBytes(n int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	d := float64(n)
	i := 0
	for d >= 1024 && i < len(units)-1 {
		d /= 1024
		i++
	}
	return fmt.Sprintf("%.2f %s", d, units[i])
}

func humanBytesF(n float64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	if n < 1024 {
		return fmt.Sprintf("%.0f B", n)
	}
	d := n
	i := 0
	for d >= 1024 && i < len(units)-1 {
		d /= 1024
		i++
	}
	return fmt.Sprintf("%.2f %s", d, units[i])
}

func handleList(conn net.Conn, metrics *Metrics) {
	request := prbp.NewCommand(prbp.REQUEST, prbp.LIST)

	sendDuration := timeIt(func() {
		n, err := conn.Write(request.Bytes())
		if err != nil {
			fmt.Println("Error sending LIST command:", err)
			return
		}
		metrics.BytesSent += n
	})
	metrics.TransmittingTime += sendDuration

	var response *prbp.Command
	var err error
	parseDuration := timeIt(func() {
		response, err = prbp.ParseCommand(conn, prbp.RESPONSE)
	})
	metrics.TransmittingTime += parseDuration

	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	metrics.BytesReceived += len(response.Bytes())

	if response.Method != prbp.LIST {
		fmt.Printf("Unexpected response method: %s (expected LIST)\n", response.Method.String())
		return
	}

	if response.PayloadSize > 0 {
		fmt.Printf("Files on server:\n%s", string(response.Payload))
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
	request := prbp.NewCommand(prbp.REQUEST, prbp.PUT).SetPayload([]byte(payload))

	sendDuration := timeIt(func() {
		n, err := conn.Write(request.Bytes())
		if err != nil {
			fmt.Println("Error sending PUT command:", err)
			return
		}
		metrics.BytesSent += n
	})
	metrics.TransmittingTime += sendDuration

	var response *prbp.Command
	parseDuration := timeIt(func() {
		response, err = prbp.ParseCommand(conn, prbp.RESPONSE)
	})
	metrics.TransmittingTime += parseDuration

	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}
	metrics.BytesReceived += len(response.Bytes())

	if response.Method != prbp.PUT {
		fmt.Printf("Unexpected response method: %s (expected PUT)\n", response.Method.String())
		return
	}

	if response.PayloadSize > 0 {
		fmt.Printf("Server response: %s\n", string(response.Payload))
	}
}

func handleQuit(conn net.Conn, metrics *Metrics) {
	request := prbp.NewCommand(prbp.REQUEST, prbp.QUIT)

	sendDuration := timeIt(func() {
		n, err := conn.Write(request.Bytes())
		if err != nil {
			fmt.Println("Error sending QUIT command:", err)
			return
		}
		metrics.BytesSent += n
	})
	metrics.TransmittingTime += sendDuration

	var response *prbp.Command
	var err error
	parseDuration := timeIt(func() {
		response, err = prbp.ParseCommand(conn, prbp.RESPONSE)
	})
	metrics.TransmittingTime += parseDuration

	if err == nil {
		metrics.BytesReceived += len(response.Bytes())
		if response.PayloadSize > 0 {
			fmt.Println("Server response:", strings.TrimSpace(string(response.Payload)))
		} else {
			fmt.Println("Server response:", response.Method.String())
		}
	} else {
		fmt.Println("Error reading response:", err)
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

	host := os.Args[1]
	port := os.Args[2]
	address := fmt.Sprintf("%s:%s", host, port)

	if err := os.MkdirAll(storage, 0755); err != nil {
		fmt.Println("Error creating client storage directory:", err)
		return
	}

	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	metrics := Metrics{StartTime: time.Now()}
	defer func() {
		metrics.EndTime = time.Now()
		dur := metrics.EndTime.Sub(metrics.StartTime).Seconds()
		fmt.Printf("\nSession Metrics:\n")
		fmt.Printf("Start Time: %s\n", metrics.StartTime.Format(time.DateTime))
		fmt.Printf("End Time: %s\n", metrics.EndTime.Format(time.DateTime))
		fmt.Printf("Duration: %.2f seconds\n", dur)
		fmt.Printf("Bytes Sent: %s\n", humanBytes(int64(metrics.BytesSent)))
		fmt.Printf("Bytes Received: %s\n", humanBytes(int64(metrics.BytesReceived)))
		txSec := metrics.TransmittingTime.Seconds()
		if txSec > 0 {
			rate := float64(metrics.BytesSent+metrics.BytesReceived) / txSec
			fmt.Printf("Throughput: %s/s\n", humanBytesF(rate))
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
