package sftpclient

import (
	"fmt"
	"io"
	"os"
)

// GetFile retreives a file from remote location
func (c *SFTPClient) GetFile(localFilename string, remoteFilename string) error {
	stats, err := c.client.Lstat(remoteFilename)
	if err != nil {
		return fmt.Errorf("Could not determine type of file 'remote:%s': %v", remoteFilename, err)
	}
	if stats.IsDir() {
		if err := os.Mkdir(localFilename, 0755); err != nil {
			return fmt.Errorf("Could not create folder 'local:%s': %v", localFilename, err)
		}
		return nil
	}

	remote, err := c.client.Open(remoteFilename)
	if err != nil {
		return fmt.Errorf("Could not open file 'remote:%s': %v", remoteFilename, err)
	}
	defer remote.Close()

	local, err := os.Create(localFilename)
	if err != nil {
		return fmt.Errorf("Could not open file 'local:%s': %v", localFilename, err)
	}
	defer local.Close()

	if _, err := io.Copy(local, remote); err != nil {
		return fmt.Errorf("Could not copy 'remote:%s' to 'local:%s': %v", remoteFilename, localFilename, err)
	}

	return nil
}

// PutFile uploads a local file to remote location
func (c *SFTPClient) PutFile(remoteFileName string, localFileName string) error {
	localFile, err := os.Open(localFileName)
	if err != nil {
		return fmt.Errorf("Could not open file 'local:%s': %v", localFileName, err)
	}
	defer localFile.Close()

	stats, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("Could not stat file 'local:%s': %v", localFileName, err)
	}

	if stats.IsDir() {
		return c.CreateDirHierarchy(remoteFileName)
	}

	remoteFile, err := c.client.Create(remoteFileName)
	if err != nil {
		return fmt.Errorf("Could not create file 'remote:%s': %v", remoteFileName, err)
	}
	defer remoteFile.Close()

	if _, err := io.Copy(remoteFile, localFile); err != nil {
		return fmt.Errorf("Could not copy data from 'local:%s' to 'remote:%s': %v", localFileName, remoteFileName, err)
	}
	remoteFile.Close()

	if err := c.client.Chmod(remoteFileName, 0644); err != nil {
		return fmt.Errorf("Could not set file permissions on 'remote:%s': %v", remoteFileName, err)
	}

	return nil
}
