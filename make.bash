#!/bin/bash
echo 'Compiling'
# export PATH="$PATH:$(go env GOPATH)/bin"
# protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./proto/service.proto
# protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./proto/ami.proto
# #GOOS=linux GOARCH=arm 
# go build
CGO_ENABLED=0 go build
if [ $? -ne 0 ]; then
	echo 'An error has occurred! Aborting the script execution...'
	exit 1
fi
# FILE=/home/rura/mnt/Linux/tulagate/tulagate
# if [ -f "$FILE" ]; then
#     echo "Mounted the server drive"
# else
#     echo "Mounting the server drive"
#     sudo mount.cifs -o username=root,password=162747 //192.168.115.23/asdu /home/rura/mnt/Linux
# fi
scp tulagate automat@10.151.50.126:/home/automat/tulagate
