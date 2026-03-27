package proxy

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Dialer interface {
	Dial(network, address string) (net.Conn, error)
}

func ListenAndServe(dialer Dialer, proxyType string, localPort int) error {
	switch proxyType {
	case "socks":
		return ListenAndServeSOCKS5(dialer, localPort)
	case "http":
		return ListenAndServeHTTP(dialer, localPort)
	default:
		return fmt.Errorf("unsupported proxy type: %s", proxyType)
	}
}

func serve(listener net.Listener, handler func(net.Conn)) {
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			handler(conn)
		}
	}()
}

func relay(a, b net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(a, b)
	}()
	go func() {
		defer wg.Done()
		io.Copy(b, a)
	}()
	wg.Wait()
}
