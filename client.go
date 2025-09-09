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
    written, err := conn.Write([]byte("PRBP PUT 128\n0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF")) 

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