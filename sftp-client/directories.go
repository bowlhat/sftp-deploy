package sftpclient

import (
	"fmt"
	"os"
	"strings"
)

// CreateDir creates a folder on remote location
func (c *SFTPClient) CreateDir(remoteFolderPath string) error {
	if _, err := c.client.Lstat(remoteFolderPath); err != nil {
		if os.IsNotExist(err) {
			if err := c.client.Mkdir(remoteFolderPath); err != nil {
				return fmt.Errorf("Could not create folder 'remote:%s': %v", remoteFolderPath, err)
			}
			if err := c.client.Chmod(remoteFolderPath, 0755); err != nil {
				return fmt.Errorf("Could not set folder permissions on 'remote:%s': %v", remoteFolderPath, err)
			}
		} else {
			return fmt.Errorf("Error finding 'remote:%s': %v", remoteFolderPath, err)
		}
	}
	return nil
}

// CreateDirHierarchy creates a folder hierarchy on remote location
func (c *SFTPClient) CreateDirHierarchy(remoteFolderPath string) error {
	parent := "."
	if strings.HasPrefix(remoteFolderPath, "/") {
		parent = "/"
		remoteFolderPath = strings.TrimPrefix(remoteFolderPath, "/")
	}

	tree := strings.Split(remoteFolderPath, "/")

	for _, dir := range tree {
		parent = strings.Join([]string{parent, dir}, "/")
		if err := c.CreateDir(parent); err != nil {
			return err
		}
	}
	return nil
}
