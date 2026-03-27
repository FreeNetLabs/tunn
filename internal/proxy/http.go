package proxy

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func ListenAndServeHTTP(dialer Dialer, localPort int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		return err
	}

	serve(listener, func(conn net.Conn) {
		handleHTTPConnection(conn, dialer)
	})

	return nil
}

func handleHTTPConnection(conn net.Conn, dialer Dialer) {
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	reader := bufio.NewReader(conn)
	req, err := http.ReadRequest(reader)
	if err != nil {
		return
	}

	if req.Method == "CONNECT" {
		handleHTTPConnect(conn, req, dialer)
	} else {
		handleHTTPRequest(conn, req, dialer)
	}
}

func handleHTTPConnect(conn net.Conn, req *http.Request, dialer Dialer) {
	host, port := splitHostPort(req.Host, 443)

	conn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))

	conn.SetReadDeadline(time.Time{})
	conn.SetWriteDeadline(time.Time{})

	address := fmt.Sprintf("%s:%d", host, port)
	remote, err := dialer.Dial("tcp", address)
	if err != nil {
		return
	}
	defer remote.Close()

	relay(conn, remote)
}

func handleHTTPRequest(conn net.Conn, req *http.Request, dialer Dialer) {
	targetHost, targetPort, targetPath := parseTarget(req)
	if targetHost == "" {
		sendHeaderError(conn, 400, "Bad Request")
		return
	}

	address := net.JoinHostPort(targetHost, strconv.Itoa(targetPort))
	remote, err := dialer.Dial("tcp", address)
	if err != nil {
		sendHeaderError(conn, 502, "Bad Gateway")
		return
	}
	defer remote.Close()

	var reqBuilder strings.Builder
	reqBuilder.WriteString(fmt.Sprintf("%s %s %s\r\n", req.Method, targetPath, req.Proto))
	for name, values := range req.Header {
		if strings.EqualFold(name, "proxy-connection") {
			continue
		}
		for _, value := range values {
			reqBuilder.WriteString(fmt.Sprintf("%s: %s\r\n", name, value))
		}
	}
	reqBuilder.WriteString("\r\n")

	remote.Write([]byte(reqBuilder.String()))
	if req.Body != nil {
		io.Copy(remote, req.Body)
		req.Body.Close()
	}

	io.Copy(conn, remote)
}

func splitHostPort(hostPort string, defaultPort int) (string, int) {
	host, portStr, err := net.SplitHostPort(hostPort)
	if err != nil {
		return hostPort, defaultPort
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return host, defaultPort
	}
	return host, port
}

func parseTarget(req *http.Request) (host string, port int, path string) {
	if req.URL.IsAbs() {
		parsedURL, err := url.Parse(req.URL.String())
		if err != nil {
			return "", 0, ""
		}
		host = parsedURL.Hostname()
		if parsedURL.Port() != "" {
			port, _ = strconv.Atoi(parsedURL.Port())
		} else {
			port = 80
			if parsedURL.Scheme == "https" {
				port = 443
			}
		}
		path = parsedURL.RequestURI()
	} else {
		if req.Host == "" {
			return "", 0, ""
		}
		host, port = splitHostPort(req.Host, 80)
		path = req.URL.RequestURI()
	}
	return
}

func sendHeaderError(conn net.Conn, code int, text string) {
	resp := fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Length: 0\r\nConnection: close\r\n\r\n", code, text)
	conn.Write([]byte(resp))
}
