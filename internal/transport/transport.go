package transport

import (
	"fmt"
	"io"

	"github.com/rzetterberg/elmobd"
)

// NewDevice creates a new OBD device based on the configuration
func NewDevice(cfg *Config) (*elmobd.Device, error) {
	var addr string
	switch cfg.Type {
	case "serial":
		addr = fmt.Sprintf("serial://%s", cfg.Address)
	case "tcp":
		addr = fmt.Sprintf("tcp://%s", cfg.Address)
	case "mock":
		addr = "mock://"
	default:
		return nil, fmt.Errorf("unsupported transport type: %s", cfg.Type)
	}
	return elmobd.NewDevice(addr, cfg.Debug)
}

// Transport represents any connection type that can be used for OBD communication
type Transport interface {
	io.ReadWriteCloser
}

// Config holds connection configuration
type Config struct {
	Type     string // "serial", "tcp", or "mock"
	Address  string // COM port or TCP address
	BaudRate int    // Only used for serial connections
	Debug    bool   // Enable debug mode
}
