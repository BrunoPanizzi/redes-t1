package main

import (
    "fmt"
    "net"
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


func handleConnection(conn net.Conn) {
    defer func() {
        conn.Close()
        fmt.Printf("[%s] Connection closed\n", conn.RemoteAddr())
    }()
    fmt.Printf("[%s] Client connected\n", conn.RemoteAddr())

    buffer := make([]byte, 32)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("[%s] Error reading from client: %v\n", conn.RemoteAddr(), err)
		return
	}

	buffer = make([]byte, 1024)

    for {
        n, err := conn.Read(buffer)
        if err != nil {
            fmt.Printf("[%s] Error reading from client: %v\n", conn.RemoteAddr(), err)
            break
        }
        fmt.Printf("[%s] Received from client: %s\n", conn.RemoteAddr(), string(buffer[:n]))
        n, err = conn.Write([]byte(fmt.Sprintf("Read your message: %s", string(buffer[:n]))))
        if err != nil {
            fmt.Printf("[%s] Error writing to client: %v\n", conn.RemoteAddr(), err)
            break
        }
        fmt.Printf("[%s] Wrote %d bytes to client\n", conn.RemoteAddr(), n)
    }
}

func main() {
    fmt.Println("Starting the tcp server on the port 8080")
    ln, err := net.Listen("tcp", ":8080")
    if err != nil {
        fmt.Println("Error starting server:", err)
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
