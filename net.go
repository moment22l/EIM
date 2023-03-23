package EIM

import "net"

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// 检查地址类型，如果不是环回则继续
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			// 检查ip是否为ipv4
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return ""
}
