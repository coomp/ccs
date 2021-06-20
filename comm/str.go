package comm

import (
	"net"
	"time"
)

// Find TODO
func Find(a []string, x string) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return -1
}

// GetLocalIpString GetLocalIpString get IPv4 address of any interface (except lo), return string which is dotted decimal notation, if fail returns ""
func GetLocalIpString() string {
	b := GetLocalIpByte()
	return net.IP(b).String()
}

// GetLocalIpByte get IPv4 address of any interface (except lo), return []byte which is dotted decimal notation,
// if fail returns nil
func GetLocalIpByte() []byte {
	addrSlice, err := net.InterfaceAddrs()
	if nil != err {
		return nil
	}
	for _, addr := range addrSlice {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip4 := ipnet.IP.To4(); nil != ip4 {
				return ip4
			}
		}
	}
	return nil
}

// Timediffer TODO
func Timediffer(start time.Time) float64 {
	return time.Now().Sub(start).Seconds()
}
