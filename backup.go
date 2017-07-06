package main

import (
	"log"

	sftpclient "github.com/bowlhat/sftp-client"

	"github.com/cheggaaa/pb"
)

func backup(c *sftpclient.SFTPClient, config backupConfig) {
	var bar *pb.ProgressBar
	if *debugging <= 1 {
		bar = pb.New(0)
	}

	files, err := c.FindAllRemoteFiles(config.From)
	if err != nil {
		log.Fatalln(err)
	}

	if *debugging <= 1 {
		bar.Total = int64(len(files))
		bar.Prefix("Backing-up... ")
		bar.Start()
	}

	saved, errors := c.BackupFiles(config.To, files)

	go func() {
		for range saved {
			bar.Increment()
		}
	}()

	errorsEncountered := false
	for err := range errors {
		errorsEncountered = true
		log.Println(err.Err)
	}
	if errorsEncountered {
		log.Fatalln("Backup failed. Quitting.")
	}

	if *debugging <= 1 {
		bar.FinishPrint("Backup complete")
	}
}
