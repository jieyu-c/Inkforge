package iputil

import "net"

// PackIP stores IPv4/v6 host bytes with IPv4 mapped to IPv6 layout (fits VARBINARY(16)).
func PackIP(host string) []byte {
	if host == "" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return nil
	}
	v6 := ip.To16()
	if v6 == nil {
		return nil
	}
	return append([]byte(nil), v6...)
}
