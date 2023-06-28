package misc

import "net"

func GetIPFromNetAddr(addr net.Addr) net.IP {
	switch raw := addr.(type) {
	case *net.UDPAddr:
		return raw.IP
	case *net.TCPAddr:
		return raw.IP
	}
	return nil
}
