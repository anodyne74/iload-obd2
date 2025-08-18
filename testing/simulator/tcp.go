package simulator

import (
	"log"
	"net"
	"time"
)

// TCPWriter implements DataWriter for TCP connections
type TCPWriter struct {
	conn net.Conn
}

// NewTCPWriter creates a new TCP simulator writer
func NewTCPWriter(conn net.Conn) DataWriter {
	return &TCPWriter{conn: conn}
}

func (w *TCPWriter) Write(data []byte) (int, error) {
	return w.conn.Write(data)
}

func (w *TCPWriter) Close() error {
	return w.conn.Close()
}

// StartTCPServer starts a TCP server for simulation
func StartTCPServer(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("Simulator listening on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	writer := NewTCPWriter(conn)
	sim := NewSimulator(writer, 100*time.Millisecond)

	log.Printf("New connection from %s", conn.RemoteAddr())

	sim.Start()
}
