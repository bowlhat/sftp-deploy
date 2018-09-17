package sftpdeployapp

import (
	"log"

	sftpclient "github.com/bowlhat/sftp-deploy/sftp-client"

	"github.com/cheggaaa/pb"
)

func upload(c *sftpclient.SFTPClient, config []sftpclient.FolderMapping) {
	var bar *pb.ProgressBar
	if *debugging <= 1 {
		bar = pb.New(0)
		bar.Total = 0
		bar.Prefix("Uploading... ")
		bar.Start()
	}

	errorsEncountered := false

	defer func() {
		if errorsEncountered {
			log.Fatalln("Upload failed. Quitting.")
		}

		if *debugging <= 1 {
			bar.FinishPrint("Upload finished")
		}
	}()

	responseChannel := make(chan sftpclient.Response)
	countChannel := make(chan int)
	done := make(chan bool)

	go c.PutFiles(config, responseChannel, countChannel, done)

	for {
		select {
		case count := <-countChannel:
			bar.Total = bar.Total + int64(count)
		case response := <-responseChannel:
			if response.Err != nil {
				log.Println("Error uploading file", response.File, response.Err)
				errorsEncountered = true
			} else if *debugging > 1 && response.File != "" {
				log.Println("Uploaded file", response.File)
			}
			if *debugging <= 1 {
				bar.Increment()
			}
		case <-done:
			return
		}
	}
}
