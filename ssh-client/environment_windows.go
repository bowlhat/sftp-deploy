package sshclient

import "os"

func CurrentUser() string {
	return os.Getenv("USERNAME")
}

func HomeDir() string {
	return os.Getenv("USERPROFILE")
}
