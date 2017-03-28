package main

import (
	"flag"
	"io/ioutil"
	"log"
	"sync/atomic"

	sftpclient "github.com/bowlhat/sftp-client"

	"strings"

	"github.com/cheggaaa/pb"
	"gopkg.in/yaml.v2"
)

type configuration struct {
	Connection struct {
		Username string
		Password string
		Hostname string
		Port     int
	}
	Backup struct {
		To   string
		From []string
	}
	Download []sftpclient.FolderMapping
	Upload   []sftpclient.FolderMapping
}

type errorResponse struct {
	Err error
}

var (
	configfilename = flag.String("config", "", "YAML configuration file")
	doBackup       = flag.Bool("backup", true, "Backup files on remote system")
	doDownload     = flag.Bool("download", false, "Download files from remote system")
	doUpload       = flag.Bool("upload", false, "Upload local files to remote system")
	debugging      = flag.Int("debug", 1, "Spew debugging info, e.g. output every file as it's touched. 0=silent, 1=progress, 2=verbose, 3=firehose")

	remoteFiles []string
)

func main() {
	flag.Parse()

	if *configfilename == "" {
		log.Fatalf("config flag indicating configuration file is required")
	}

	configfile, err := ioutil.ReadFile(*configfilename)
	if err != nil {
		log.Fatalf("could not read config file: %v", err)
	}

	config := configuration{}
	if err := yaml.Unmarshal(configfile, &config); err != nil {
		log.Fatalf("YAML error: %v", err)
	}

	upload := *doUpload
	backup := *doBackup
	download := *doDownload

	// start ssh client on tcp connection
	sftp, err := sftpclient.New(
		config.Connection.Hostname,
		config.Connection.Port,
		config.Connection.Username,
		config.Connection.Password)

	if err != nil {
		log.Fatal(err)
	}

	var bar *pb.ProgressBar
	if *debugging <= 1 {
		bar = pb.New(0)
	}

	if backup == true {
		files, err := sftp.FindAllRemoteFiles(config.Backup.From)
		if err != nil {
			log.Fatalln(err)
		}

		if *debugging <= 1 {
			bar.Total = 0
			bar.Prefix("Backing-up... ")
			bar.Start()
		}

		saved, errors := sftp.BackupFiles(config.Backup.To, files)

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

	if download == true {
		downloadedFile := make(chan bool)
		downloadDone := make(chan bool)
		errorsChannel := make(chan errorResponse)

		if *debugging <= 1 {
			bar.Total = 0
			bar.Prefix("Downloading... ")
			bar.Start()
		}

		for _, folder := range config.Download {
			go func(f sftpclient.FolderMapping) {
				defer func() {
					downloadDone <- true
				}()

				media, err := sftp.FindAllRemoteFiles([]string{f.Remote})
				if err != nil {
					errorsChannel <- errorResponse{err}
					return
				}
				atomic.AddInt64(&bar.Total, int64(len(media)))
				for _, remotefile := range media {
					trimmed := strings.TrimPrefix(remotefile, f.Remote)
					localfile := strings.Join([]string{f.Local, trimmed}, "/")
					if err := sftp.GetFile(localfile, remotefile); err != nil {
						errorsChannel <- errorResponse{err}
					}
					downloadedFile <- true
				}
			}(folder)
		}

		go func() {
			for range downloadedFile {
				<-downloadedFile
				bar.Increment()
			}
		}()

		go func() {
			for err := range errorsChannel {
				log.Println(err)
			}
		}()

		for range config.Download {
			<-downloadDone
		}
		close(downloadDone)
		close(downloadedFile)
		close(errorsChannel)

		if *debugging <= 1 {
			bar.FinishPrint("Download complete")
		}
	}

	if upload == true {
		if *debugging <= 1 {
			bar.Prefix("Uploading... ")
			bar.Start()
		}

		errorChannel, countChannel, copiedCountChannel := sftp.Upload(config.Upload)

		go func() {
			for range countChannel {
				atomic.AddInt64(&bar.Total, 1)
			}
		}()
		go func() {
			for range copiedCountChannel {
				bar.Increment()
			}
		}()

		for response := range errorChannel {
			log.Println(response.Err)
		}

		if *debugging <= 1 {
			bar.FinishPrint("Upload finished")
		}
	}
}
