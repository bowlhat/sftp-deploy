package sftpdeployapp

import (
	"log"

	sftpclient "github.com/bowlhat/sftp-deploy/sftp-client"

	"github.com/cheggaaa/pb"
)

func backup(c *sftpclient.SFTPClient, config backupConfig) {
	var bar *pb.ProgressBar
	if *debugging <= 1 {
		bar = pb.New(0)
		bar.Total = 0
		bar.Prefix("Backing-up... ")
		bar.Start()
	}

	errorsEncountered := false

	defer func() {
		if errorsEncountered {
			log.Fatalln("Backup failed. Quitting.")
		}

		if *debugging <= 1 {
			bar.FinishPrint("Backup complete")
		}
	}()

	files, err := c.FindAllRemoteFiles(config.From)
	if err != nil {
		log.Fatalln(err)
	}

	bar.Total = int64(len(files))

	responseChannel := make(chan sftpclient.Response)

	go c.BackupFiles(config.To, files, responseChannel)

	for response := range responseChannel {
		if response.Err != nil {
			log.Println("Error backing-up file", response.File, response.Err)
			errorsEncountered = true
		} else if *debugging > 1 && response.File != "" {
			log.Println("Backed-up file", response.File)
		}
		if *debugging <= 1 {
			bar.Increment()
		}
	}
}
