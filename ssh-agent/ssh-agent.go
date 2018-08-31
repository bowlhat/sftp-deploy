package sshagent

import (
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func New() func() ([]ssh.Signer, error) {
	if aconn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return agent.NewClient(aconn).Signers
	}

	return nil
}
