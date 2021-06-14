package comm

import "net"

func Find(a []string, x string) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return -1
}
// get IPv4 address of any interface (except lo), return string which is dotted decimal notation,
// if fail returns ""
func GetLocalIpString() string {
	b := GetLocalIpByte()
	return net.IP(b).String()
}

// get IPv4 address of any interface (except lo), return []byte which is dotted decimal notation,
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
