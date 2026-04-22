package transport

import (
	"fmt"
	"io"

	"github.com/rzetterberg/elmobd"
)

// NewDevice creates a new OBD device based on the configuration
func NewDevice(cfg *Config) (Device, error) {
	var addr string
	switch cfg.Type {
	case "serial":
		addr = fmt.Sprintf("serial://%s", cfg.Address)
		return newELMDevice(addr, cfg.Debug)
	case "tcp":
		addr = fmt.Sprintf("tcp://%s", cfg.Address)
		return newELMDevice(addr, cfg.Debug)
	case "mock":
		return newELMDevice("test://", cfg.Debug)
	case "ble":
		return NewBLEDevice(cfg)
	default:
		return nil, fmt.Errorf("unsupported transport type: %s", cfg.Type)
	}
}

func newELMDevice(addr string, debug bool) (Device, error) {
	dev, err := elmobd.NewDevice(addr, debug)
	if err != nil {
		return nil, err
	}
	return &ELMDevice{device: dev}, nil
}

// Device is the minimal contract the app needs for querying OBD commands.
type Device interface {
	RunOBDCommand(cmd elmobd.OBDCommand) (elmobd.OBDCommand, error)
}

// ELMDevice wraps github.com/rzetterberg/elmobd.Device.
type ELMDevice struct {
	device *elmobd.Device
}

func (d *ELMDevice) RunOBDCommand(cmd elmobd.OBDCommand) (elmobd.OBDCommand, error) {
	return d.device.RunOBDCommand(cmd)
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
	BLE      BLEConfig
}

// BLEConfig holds Linux BLE GATT settings used for MX201-style adapters.
type BLEConfig struct {
	DeviceName         string
	MACAddress         string
	ServiceUUID        string
	WriteCharUUID      string
	NotifyCharUUID     string
	ScanTimeoutSeconds int
	CommandTimeoutMS   int
}
