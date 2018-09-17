package sftpclient

import (
	"github.com/pkg/sftp"

	sshclient "github.com/bowlhat/sftp-deploy/ssh-client"
)

// SFTPClient is a proxy to sftp.Client
type SFTPClient struct {
	client     *sftp.Client
	connection *sshclient.SSHConnection
}

// Response is a reponse which contains a filename or error
type Response struct {
	File string
	Err  error
}

// FolderMapping maps local to remote folders
type FolderMapping struct {
	Local  string
	Remote string
}

// New SFTP Connection
func New(hostname string, port int, username string, password string) (client *SFTPClient, err error) {
	ssh, err := sshclient.New(hostname, port, username, password)
	if err != nil {
		return nil, err
	}

	sftpClient, err := sftp.NewClient(ssh.Client)
	if err != nil {
		ssh.Close()
		return nil, err
	}

	return &SFTPClient{client: sftpClient, connection: ssh}, nil
}

// Close the SFTP session
func (c *SFTPClient) Close() error {
	return c.connection.Close()
}
