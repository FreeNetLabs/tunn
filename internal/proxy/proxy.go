package proxy

import (
	"io"
	"net"
	"sync"
)

type Dialer interface {
	Dial(network, address string) (net.Conn, error)
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
