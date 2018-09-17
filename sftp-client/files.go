package sftpclient

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/kr/fs"
)

// GetFiles retreives all files from a remote location
func (c *SFTPClient) GetFiles(folders []FolderMapping, response chan<- Response, count chan<- int, done chan<- bool) {
	defer func() {
		close(response)
		close(count)
		close(done)
	}()

	for _, folder := range folders {
		media, err := c.FindAllRemoteFiles([]string{folder.Remote})
		if err != nil {
			log.Println(err)
			return
		}

		count <- len(media)

		for _, remoteFile := range media {
			trimmed := strings.TrimPrefix(remoteFile, folder.Remote)
			localFile := strings.Join([]string{folder.Local, trimmed}, "/")

			err := c.GetFile(localFile, remoteFile)
			if err != nil {
				response <- Response{File: remoteFile, Err: err}
				continue
			}

			response <- Response{File: remoteFile, Err: nil}
		}
	}

	done <- true
}

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

// PutFiles uploads all local files to remote location
func (c *SFTPClient) PutFiles(folders []FolderMapping, response chan<- Response, count chan<- int, done chan<- bool) {
	defer func() {
		close(response)
		close(count)
		close(done)
	}()

	for _, folder := range folders {
		if _, err := os.Lstat(folder.Local); err != nil {
			continue
		}

		fullLocalPath, err := filepath.EvalSymlinks(folder.Local)
		if err != nil {
			continue
		}

		files := []string{}

		walker := fs.Walk(fullLocalPath)
		for walker.Step() {
			if err := walker.Err(); err != nil {
				log.Println(err)
				continue
			}

			p := walker.Path()
			if p == fullLocalPath {
				continue
			}

			files = append(files, p)
		}

		count <- len(files)

		for _, file := range files {
			remoteFileName := strings.TrimPrefix(file, folder.Local)
			remoteFileName = strings.TrimPrefix(remoteFileName, "/")
			remoteFileName = strings.Join([]string{folder.Remote, remoteFileName}, "/")

			remoteDir := path.Dir(remoteFileName)
			if err := c.CreateDirHierarchy(remoteDir); err != nil {
				response <- Response{File: "", Err: err}
				continue
			}

			if err := c.PutFile(remoteFileName, file); err != nil {
				response <- Response{File: file, Err: err}
				continue
			}

			response <- Response{File: file, Err: nil}
		}
	}

	done <- true
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
