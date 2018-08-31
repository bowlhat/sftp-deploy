package sftpdeployapp

import (
	"flag"
	"io/ioutil"
	"log"

	sftpclient "github.com/bowlhat/sftp-deploy/sftp-client"

	"gopkg.in/yaml.v2"
)

type backupConfig struct {
	To   string
	From []string
}
type configuration struct {
	Connection struct {
		Username string
		Password string
		Hostname string
		Port     int
	}
	Backup   backupConfig
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

// Main is the entry to the sftp-deploy application
func Main() {
	flag.Parse()

	if *configfilename == "" {
		log.Fatalf("-config flag is required, see: -help")
	}

	configfile, err := ioutil.ReadFile(*configfilename)
	if err != nil {
		log.Fatalf("could not read config file: %v", err)
	}

	config := configuration{}
	if err := yaml.Unmarshal(configfile, &config); err != nil {
		log.Fatalf("YAML error: %v", err)
	}

	// start ssh client on tcp connection
	sftp, err := sftpclient.New(
		config.Connection.Hostname,
		config.Connection.Port,
		config.Connection.Username,
		config.Connection.Password)

	if err != nil {
		log.Fatalf("SSH Conenction error: %v", err)
	}

	if *doBackup == true {
		backup(sftp, config.Backup)
	}

	if *doDownload == true {
		download(sftp, config.Download)
	}

	if *doUpload == true {
		upload(sftp, config.Upload)
	}
}
