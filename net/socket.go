package net

import (
	"fmt"
	"net"
	"syscall"
)

// Socket pack up socket API of os.
type Socket struct {
	Fd int
}

func NewSocket() (*Socket, error) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, fmt.Errorf("create socket err: %s", err)
	}

	return &Socket{Fd: fd}, nil
}

func (s *Socket) Read(bytes []byte) (int, error) {
	if len(bytes) == 0 {
		return 0, nil
	}
	n, err := syscall.Read(s.Fd, bytes)
	if err != nil {
		return 0, fmt.Errorf("read socket err: %s", err)
	}
	return n, nil
}

func (s *Socket) Write(bytes []byte) (int, error) {
	n, err := syscall.Write(s.Fd, bytes)
	if err != nil {
		return 0, fmt.Errorf("write socket err: %s", err)
	}
	return n, nil
}

func (s *Socket) Close() error {
	err := syscall.Close(s.Fd)
	if err != nil {
		return fmt.Errorf("close socket err: %s", err)
	}
	return nil
}

func (s *Socket) String() string {
	return fmt.Sprintf("socket{ fd: %d }", s.Fd)
}

func Listen(ip string, port int) (*Socket, error) {
	socket, err := NewSocket()
	if err != nil {
		return nil, fmt.Errorf("create socket err: (%s)", err)
	}
	parseIP := net.ParseIP(ip)
	addr := &syscall.SockaddrInet4{Port: port}
	copy(addr.Addr[:], parseIP)

	if err = syscall.Bind(socket.Fd, addr); err != nil {
		return nil, fmt.Errorf("bind socket to addr err: (%s)", err)
	}

	if err = syscall.Listen(socket.Fd, syscall.SOMAXCONN); err != nil {
		return nil, fmt.Errorf("socket listen err: (%s)", err)
	}

	return socket, nil
}
