package escpos

import (
	"fmt"
	"net"
	"time"
)

type NetworkPrinter struct {
	IPAddress string
	Port      int
	Timeout   time.Duration
}

func NewNetworkPrinter(ipAddress string, port int) *NetworkPrinter {
	if port == 0 {
		port = 9100
	}
	return &NetworkPrinter{
		IPAddress: ipAddress,
		Port:      port,
		Timeout:   10 * time.Second,
	}
}

func (p *NetworkPrinter) Print(data []byte) error {
	addr := fmt.Sprintf("%s:%d", p.IPAddress, p.Port)
	conn, err := net.DialTimeout("tcp", addr, p.Timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to printer %s: %w", addr, err)
	}
	defer conn.Close()

	if err := conn.SetWriteDeadline(time.Now().Add(p.Timeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send data to printer: %w", err)
	}

	return nil
}

func (p *NetworkPrinter) CheckStatus() error {
	addr := fmt.Sprintf("%s:%d", p.IPAddress, p.Port)
	conn, err := net.DialTimeout("tcp", addr, p.Timeout)
	if err != nil {
		return fmt.Errorf("printer not reachable: %w", err)
	}
	defer conn.Close()
	return nil
}
