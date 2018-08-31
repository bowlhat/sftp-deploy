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

	errorsEncountered := false
	saved, errors, done := c.BackupFiles(config.To, files)

	defer func() {
		if errorsEncountered {
			log.Fatalln("Backup failed. Quitting.")
		}

		if *debugging <= 1 {
			bar.FinishPrint("Backup complete")
		}
	}()

	for {
		select {
		case <-saved:
			if *debugging <= 1 {
				bar.Increment()
			}
		case err := <-errors:
			errorsEncountered = true
			log.Println(err.Err)
		case <-done:
			return
		}
	}
}
