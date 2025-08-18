package simulator

import (
	"github.com/tarm/serial"
)

// SerialWriter implements DataWriter for serial ports
type SerialWriter struct {
	port *serial.Port
}

// NewSerialWriter creates a new serial simulator writer
func NewSerialWriter(portName string, baud int) (DataWriter, error) {
	config := &serial.Config{
		Name: portName,
		Baud: baud,
	}

	port, err := serial.OpenPort(config)
	if err != nil {
		return nil, err
	}

	return &SerialWriter{port: port}, nil
}

func (w *SerialWriter) Write(data []byte) (int, error) {
	return w.port.Write(data)
}

func (w *SerialWriter) Close() error {
	return w.port.Close()
}
