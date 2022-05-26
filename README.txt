create new go project:
$ go mod init mindnight-koder.net/l1vmgodata

build:
$ go build

list os/arch:
$ go tool dist list

build using os/arch 
$ GOOS=windows GOARCH=amd64 go build

run remote on server via ssh:
$ nohup ./l1vmgodata.sh &
