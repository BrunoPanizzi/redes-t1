package prbp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

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

type Command struct {
	Type        CommandType
	Method      Method
	PayloadSize int
	Payload     []byte
}

func (m Method) String() string {
	switch m {
	case LIST:
		return "LIST"
	case PUT:
		return "PUT"
	case QUIT:
		return "QUIT"
	default:
		return "UNKNOWN"
	}
}

func (c *Command) String() string {
	return fmt.Sprintf("Type: %v, Method: %v, PayloadSize: %d, Payload: %s{}",
		c.Type, c.Method, c.PayloadSize, "lots of things")
}

func (c *Command) Bytes() []byte {
	header := fmt.Sprintf("PRBP %s %d\n", c.Method.String(), c.PayloadSize)
	return append([]byte(header), c.Payload...)
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

func ParseCommand(r io.Reader, t CommandType) (*Command, error) {
	fmt.Printf("Parsing command of type %v\n", t)
	reader := bufio.NewReader(r)

	header, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(string(header), " ", 4)
	fmt.Printf("Header parts: %v\n", parts)

	if parts[0] != "PRBP" {
		return nil, fmt.Errorf("invalid protocol")
	}

	method, err := ParseMethod(parts[1])
	if err != nil {
		return nil, err
	}

	var payloadSize int
	if len(parts) > 2 {
		payloadSize, err = strconv.Atoi(strings.Trim(parts[2], "\n"))
		if err != nil {
			return nil, err
		}
	}

	cmd := &Command{
		Type:        t,
		Method:      method,
		PayloadSize: payloadSize,
		Payload:     make([]byte, payloadSize),
	}

	if cmd.PayloadSize > 0 {
		_, err := io.ReadFull(reader, cmd.Payload)
		if err != nil {
			return nil, err
		}
	}

	return cmd, nil
}

func NewCommand(t CommandType, m Method) *Command {
	return &Command{
		Type:   t,
		Method: m,
	}
}

func (c *Command) SetPayload(payload []byte) *Command {
	c.Payload = payload
	c.PayloadSize = len(payload)
	return c
}
