package transport

import (
	"net"
)

// TCPConnection implements io.ReadWriteCloser for TCP connections
type TCPConnection struct {
	conn net.Conn
}

// NewTCPConnection creates a new TCP connection to the specified address
func NewTCPConnection(addr string) (*TCPConnection, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &TCPConnection{conn: conn}, nil
}

func (t *TCPConnection) Read(p []byte) (n int, err error) {
	return t.conn.Read(p)
}

func (t *TCPConnection) Write(p []byte) (n int, err error) {
	return t.conn.Write(p)
}

func (t *TCPConnection) Close() error {
	return t.conn.Close()
}
