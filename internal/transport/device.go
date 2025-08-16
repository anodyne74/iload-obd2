package transport

import (
	"fmt"

	"github.com/rzetterberg/elmobd"
)

// NewConnection creates a new connection based on the configuration
func NewConnection(cfg *Config) (Transport, error) {
	switch cfg.Type {
	case "tcp":
		return NewTCPConnection(cfg.Address)
	case "serial":
		dev, err := elmobd.NewDevice(cfg.Address, true) // true = debug mode
		if err != nil {
			return nil, fmt.Errorf("failed to create serial connection: %v", err)
		}
		return &serialTransport{device: dev}, nil
	case "mock":
		// TODO: Implement mock connection
		return nil, fmt.Errorf("mock connection not implemented")
	default:
		return nil, fmt.Errorf("unsupported transport type: %s", cfg.Type)
	}
}

// serialTransport wraps the elmobd.Device to implement our Transport interface
type serialTransport struct {
	device *elmobd.Device
}

func (s *serialTransport) Read(p []byte) (n int, err error) {
	// Implementation depends on elmobd.Device's actual API
	// TODO: Implement based on the device's actual reading mechanism
	return 0, fmt.Errorf("not implemented")
}

func (s *serialTransport) Write(p []byte) (n int, err error) {
	// Implementation depends on elmobd.Device's actual API
	// TODO: Implement based on the device's actual writing mechanism
	return 0, fmt.Errorf("not implemented")
}

func (s *serialTransport) Close() error {
	// Implementation depends on elmobd.Device's actual API
	// TODO: Implement based on the device's actual close mechanism
	return nil
}
