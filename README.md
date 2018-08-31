# sftp-deploy

This is a deployment script for sftp-only sites. It is written in Go. Originally developed with specifics for [Bang Communications](http://www.bang-on.net/), and subsequently generalised and released publicly.

## Installation
To install you first need to get the Go Compiler from either your OS package repository or the [Go Homepage](https://golang.org/).

Once you've got Go installed you need to set aside a folder for all Go projects to reside within. This folder needs to be declared in an environment variable called GOPATH. (It's advisable to set this somewhere that it is permanent.)

Now that the GOPATH is defined you can carry-on and run `go get github.com/bowlhat/sftp-deploy` which will fetch the code and all dependencies and compile it all into a single binary which will be stored at `$GOPATH/bin/sftp-deploy`.

## Configuration
Configuration is handled via a YAML file which must be specified using the `-config` commandline flag.

The configuration format is as follows:

```yaml
connection:
  username: ""
  password: ""
  hostname: "127.0.0.1"
  port: 22
backup:
  to: "local/path/to/backup/folder"
  from:
    - "remote/folder/path"
    - "remote/folder/path2"
    - ...
download:
  - local: "local/destination/folder1"
    remote: "remote/source/folder1"
  - local: "local/destination/folder2"
    remote: "remote/source/folder2"
  - ...
upload:
  - local: "local/source/folder1"
    remote: "remote/destination/folder1"
  - local: "local/source/folder2"
    remote: "remote/destination/folder2"
  - ...
```

## Running
To get a list of available flags, their meanings and default values run `./sftp-deploy -help`.

When deploying to a live site it is recommended to use the `-backup=true` flag, which will create a time-stamped .tar.gz archive of the files that the program removes. This file can be found in the backup folder specified in the configuration file, which will be created if it doesn't already exist. As the backup is time-stamped there is no need to remove an old backup before running the program as it won't get overwritten.

Ideally you will have an ssh-agent running with a valid key loaded which is accepted by the sftp server. (On Windows the supported ssh-agent is Pageant which is available from [the PuTTY project](http://www.chiark.greenend.org.uk/~sgtatham/putty/download.html)) However you can use the configuration file to specify a password if required. The program will NOT prompt for a password interactively so if password-based authentication is required then you MUST supply the password in the configuration.

## Our Packages

- [SSH Agent abstraction](https://github.com/bowlhat/sftp-deploy/ssh-agent)
- [SSH Client abstraction](https://github.com/bowlhat/sftp-deploy/ssh-client)
- [SFTP Client abstraction](https://github.com/bowlhat/sftp-deploy/sftp-client)

## Related Code
These packages are included automatically by the `go get` command. They're documented here for completeness.

- [Progressbars](https://github.com/cheggaaa/pb) are powered by cheggaaa's progressbar project
