package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync/atomic"

	sftpclient "github.com/bowlhat/sftp-client"

	"github.com/cheggaaa/pb"
	"github.com/kr/fs"
)

func foundFileChannels(channels []chan sftpclient.FileResponse) func(sftpclient.FileResponse) {
	return func(res sftpclient.FileResponse) {
		for _, channel := range channels {
			channel <- res
		}
	}
}

// Upload and synchronize folder trees
func uploadProcessor(c *sftpclient.SFTPClient, folders []sftpclient.FolderMapping) (found <-chan sftpclient.FileResponse, copied <-chan sftpclient.FileResponse, done <-chan bool) {
	doneChannel := make(chan bool)
	bubbleDoneChannel := make(chan bool)

	foundFilesResponseChannel := make(chan sftpclient.FileResponse)
	copiedFilesResponseChannel := make(chan sftpclient.FileResponse)

	for _, folder := range folders {
		channel := make(chan sftpclient.FileResponse)

		channels := []chan sftpclient.FileResponse{foundFilesResponseChannel, channel}
		foundFile := foundFileChannels(channels)

		go func(f sftpclient.FolderMapping) {
			if _, err := os.Lstat(f.Local); err != nil {
				foundFile(sftpclient.FileResponse{File: "", Err: err})
				return
			}

			fullLocalPath, err := filepath.EvalSymlinks(f.Local)
			if err != nil {
				foundFile(sftpclient.FileResponse{File: "", Err: err})
				return
			}

			walker := fs.Walk(fullLocalPath)

			for walker.Step() {
				if err := walker.Err(); err != nil {
					log.Println(err)
					foundFile(sftpclient.FileResponse{File: "", Err: err})
					continue
				}

				p := walker.Path()
				if p == fullLocalPath {
					continue
				}

				foundFile(sftpclient.FileResponse{File: p, Err: nil})
			}

			close(channel)
		}(folder)

		go func(f sftpclient.FolderMapping) {
			for response := range channel {
				if response.Err != nil {
					copiedFilesResponseChannel <- sftpclient.FileResponse{File: "", Err: response.Err}
					continue
				}

				remoteFileName := strings.TrimPrefix(response.File, f.Local)
				remoteFileName = strings.TrimPrefix(remoteFileName, "/")
				remoteFileName = strings.Join([]string{f.Remote, remoteFileName}, "/")

				remoteDir := path.Dir(remoteFileName)
				if err := c.CreateDirHierarchy(remoteDir); err != nil {
					copiedFilesResponseChannel <- sftpclient.FileResponse{File: "", Err: err}
					continue
				}

				if err := c.PutFile(remoteFileName, response.File); err != nil {
					copiedFilesResponseChannel <- sftpclient.FileResponse{File: "", Err: err}
					continue
				}
				copiedFilesResponseChannel <- sftpclient.FileResponse{File: response.File, Err: nil}
			}
			doneChannel <- true
		}(folder)
	}

	go func() {
		for range folders {
			<-doneChannel
		}
		bubbleDoneChannel <- true
		close(copiedFilesResponseChannel)
		close(foundFilesResponseChannel)
		close(doneChannel)
		close(bubbleDoneChannel)
	}()

	return foundFilesResponseChannel, copiedFilesResponseChannel, bubbleDoneChannel
}

func upload(c *sftpclient.SFTPClient, config []sftpclient.FolderMapping) {
	var bar *pb.ProgressBar
	if *debugging <= 1 {
		bar = pb.New(0)
		bar.Total = 0
		bar.Prefix("Uploading... ")
		bar.Start()
		defer bar.FinishPrint("Upload finished")
	} else {
		defer log.Println("Finished")
	}

	foundFiles, copiedFiles, done := uploadProcessor(c, config)

	for {
		select {
		case file := <-foundFiles:
			if *debugging > 1 {
				log.Println("Found file for upload", file.File)
			} else {
				atomic.AddInt64(&bar.Total, 1)
			}
		case file := <-copiedFiles:
			if file.Err != nil {
				log.Println(fmt.Errorf("Received error copying: %v", file.Err))
			} else if *debugging > 1 {
				log.Println("Uploaded file", file.File)
			}

			if *debugging <= 1 {
				bar.Increment()
			}
		case <-done:
			return
		}
	}
}
