//go:build !linux

package transport

import "fmt"

// NewBLEDevice is only available on Linux in this project.
func NewBLEDevice(cfg *Config) (Device, error) {
	return nil, fmt.Errorf("ble transport is currently only supported on linux")
}
