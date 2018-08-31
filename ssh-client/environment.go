package sshclient

import "os"

// CurrentUser Get the user name of the current system user
func CurrentUser() string {
	return os.Getenv("USER")
}

// HomeDir Get the path to the current system user's home directory
func HomeDir() string {
	return os.Getenv("HOME")
}
