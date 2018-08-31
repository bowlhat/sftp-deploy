package sshagent

import (
	"github.com/davidmz/go-pageant"
	"golang.org/x/crypto/ssh"
)

func New() func() ([]ssh.Signer, error) {
	if pageant.Available() {
		return pageant.New().Signers
	}

	return nil
}
