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

	"github.com/FreeNetLabs/tunn/pkg/utils"
)

type HTTP struct {
	server *Server
}

func NewHTTP(ssh SSHClient) *HTTP {
	return &HTTP{
		server: NewServer(ssh),
	}
}

func (h *HTTP) Start(localPort int) error {
	return h.server.StartProxy("HTTP", localPort, h.handleClient)
}

func (h *HTTP) handleClient(clientConn net.Conn) {
	h.server.HandleClientWithTimeout(clientConn, "HTTP", 30*time.Second, func() {
		reader := bufio.NewReader(clientConn)
		req, err := http.ReadRequest(reader)
		if err != nil {
			fmt.Printf("✗ Error reading HTTP request: %v\n", err)
			h.sendError(clientConn, 400, "Bad Request")
			return
		}

		if req.Method == "CONNECT" {
			h.handleConnect(clientConn, req)
		} else {
			h.handleRequest(clientConn, req)
		}
	})
}

func (h *HTTP) handleConnect(clientConn net.Conn, req *http.Request) {
	host, portInt, err := utils.ParseHostPort(req.Host, 443)
	if err != nil {
		fmt.Printf("✗ Invalid host in CONNECT request: %v\n", err)
		h.sendError(clientConn, 400, "Bad Request")
		return
	}

	fmt.Printf("→ HTTP CONNECT request to %s:%d\n", host, portInt)

	response := "HTTP/1.1 200 Connection established\r\n\r\n"
	if _, err := clientConn.Write([]byte(response)); err != nil {
		fmt.Printf("✗ Error sending CONNECT response: %v\n", err)
		return
	}

	fmt.Printf("✓ HTTP CONNECT tunnel established to %s:%d\n", host, portInt)
	h.server.OpenSSHChannel(clientConn, host, portInt)
}

func (h *HTTP) handleRequest(clientConn net.Conn, req *http.Request) {
	targetHost, targetPort, targetPath, err := h.parseTarget(req)
	if err != nil {
		fmt.Printf("✗ Error parsing HTTP target: %v\n", err)
		h.sendError(clientConn, 400, "Bad Request")
		return
	}

	fmt.Printf("→ HTTP %s request to %s:%d%s\n", req.Method, targetHost, targetPort, targetPath)

	address := net.JoinHostPort(targetHost, strconv.Itoa(targetPort))
	sshConn, err := h.server.ssh.Dial("tcp", address)
	if err != nil {
		fmt.Printf("✗ Failed to open SSH channel for HTTP request: %v\n", err)
		h.sendError(clientConn, 502, "Bad Gateway")
		return
	}
	defer sshConn.Close()

	if err := h.forwardRequest(sshConn, req, targetPath); err != nil {
		fmt.Printf("✗ Error forwarding HTTP request: %v\n", err)
		h.sendError(clientConn, 502, "Bad Gateway")
		return
	}

	h.forwardResponse(clientConn, sshConn)
}

func (h *HTTP) parseTarget(req *http.Request) (host string, port int, path string, err error) {
	if req.URL.IsAbs() {
		parsedURL, err := url.Parse(req.URL.String())
		if err != nil {
			return "", 0, "", err
		}

		host = parsedURL.Hostname()
		if parsedURL.Port() != "" {
			port, err = strconv.Atoi(parsedURL.Port())
			if err != nil {
				return "", 0, "", fmt.Errorf("invalid port in URL: %s", parsedURL.Port())
			}
		} else {
			port = 80
			if parsedURL.Scheme == "https" {
				port = 443
			}
		}
		path = parsedURL.RequestURI()
	} else {
		if req.Host == "" {
			return "", 0, "", fmt.Errorf("no Host header in HTTP request")
		}

		host, port, err = utils.ParseHostPort(req.Host, 80)
		if err != nil {
			return "", 0, "", fmt.Errorf("invalid Host header: %v", err)
		}
		path = req.URL.RequestURI()
	}

	return host, port, path, nil
}

func (h *HTTP) forwardRequest(sshConn net.Conn, req *http.Request, targetPath string) error {
	var requestBuilder strings.Builder

	requestBuilder.WriteString(fmt.Sprintf("%s %s %s\r\n", req.Method, targetPath, req.Proto))

	for name, values := range req.Header {
		if strings.ToLower(name) == "proxy-connection" {
			continue
		}
		for _, value := range values {
			requestBuilder.WriteString(fmt.Sprintf("%s: %s\r\n", name, value))
		}
	}

	requestBuilder.WriteString("\r\n")

	_, err := sshConn.Write([]byte(requestBuilder.String()))
	if err != nil {
		return err
	}

	if req.Body != nil {
		_, err = io.Copy(sshConn, req.Body)
		req.Body.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *HTTP) forwardResponse(clientConn net.Conn, sshConn net.Conn) {
	_, err := io.Copy(clientConn, sshConn)
	if err != nil && err != io.EOF {
		fmt.Printf("✗ Error forwarding HTTP response: %v\n", err)
	}
}

func (h *HTTP) sendError(clientConn net.Conn, statusCode int, statusText string) {
	response := fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Length: 0\r\nConnection: close\r\n\r\n", statusCode, statusText)
	clientConn.Write([]byte(response))
}
