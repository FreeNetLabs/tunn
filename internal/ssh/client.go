package ssh

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/FreeNetLabs/tunn/internal/config"
)

func NewClient(conn net.Conn, cfg *config.Config) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: cfg.Auth.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(cfg.Auth.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
		BannerCallback: func(message string) error {
			msg := strings.TrimSpace(message)
			if msg != "" {
				log.Printf("ssh banner: %s", msg)
			}
			return nil
		},
	}

	log.Printf("establishing ssh connection for user %s...", cfg.Auth.Username)
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, "tcp", config)
	if err != nil {
		return nil, fmt.Errorf("ssh conn err: %w", err)
	}

	return ssh.NewClient(sshConn, chans, reqs), nil
}
