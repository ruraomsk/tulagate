#!/bin/bash
echo 'Compiling'
export PATH="$PATH:$(go env GOPATH)/bin"
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./proto/service.proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./proto/ami.proto
#GOOS=linux GOARCH=arm 
go build
if [ $? -ne 0 ]; then
	echo 'An error has occurred! Aborting the script execution...'
	exit 1
fi
#echo 'Copy kuda to device'
#scp kuda admin@192.168.115.29:/home/admin
#scp test.bin admin@192.168.115.29:/home/admin
