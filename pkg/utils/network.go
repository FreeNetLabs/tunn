package utils

import (
	"fmt"
	"net"
	"strconv"
)

func ParseHostPort(hostPort string, defaultPort int) (string, int, error) {
	host, portStr, err := net.SplitHostPort(hostPort)
	if err != nil {
		return hostPort, defaultPort, nil
	}

	switch portStr {
	case "https":
		return host, 443, nil
	case "http":
		return host, 80, nil
	default:
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return "", 0, fmt.Errorf("invalid port: %s", portStr)
		}
		return host, port, nil
	}
}
