package sftpdeployapp

import (
	"log"
	"strings"
	"sync/atomic"

	sftpclient "github.com/bowlhat/sftp-deploy/sftp-client"

	"github.com/cheggaaa/pb"
)

func download(c *sftpclient.SFTPClient, config []sftpclient.FolderMapping) {
	var bar *pb.ProgressBar
	if *debugging <= 1 {
		bar = pb.New(0)
	}

	downloadedFile := make(chan bool)
	downloadDone := make(chan bool)

	if *debugging <= 1 {
		bar.Total = 0
		bar.Prefix("Downloading... ")
		bar.Start()
	}

	for _, folder := range config {
		go func(f sftpclient.FolderMapping) {
			defer func() {
				downloadDone <- true
			}()

			media, err := c.FindAllRemoteFiles([]string{f.Remote})
			if err != nil {
				log.Println(err)
				return
			}
			atomic.AddInt64(&bar.Total, int64(len(media)))
			for _, remotefile := range media {
				trimmed := strings.TrimPrefix(remotefile, f.Remote)
				localfile := strings.Join([]string{f.Local, trimmed}, "/")
				if err := c.GetFile(localfile, remotefile); err != nil {
					log.Println(err)
				}
				downloadedFile <- true
			}
		}(folder)
	}

	go func() {
		for range downloadedFile {
			if *debugging <= 1 {
				bar.Increment()
			}
		}
	}()

	for range config {
		<-downloadDone
	}
	close(downloadDone)
	close(downloadedFile)

	if *debugging <= 1 {
		bar.FinishPrint("Download complete")
	}
}
