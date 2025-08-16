package transport

import (
	"io"
)

// Transport represents any connection type that can be used for OBD communication
type Transport interface {
	io.ReadWriteCloser
}

// Config holds connection configuration
type Config struct {
	Type     string // "serial", "tcp", or "mock"
	Address  string // COM port or TCP address
	BaudRate int    // Only used for serial connections
}
