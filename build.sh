#!/bin/bash
go get bpm
go install bpm
if [ -f /usr/local/bin/bpm ]; then
    rm /usr/local/bin/bpm
fi
echo Creating symbolic link
ln -s `pwd`/bin/bpm /usr/local/bin/bpm

echo Creating linux version
env GOPATH=`pwd` GOOS=linux GOARCH=amd64 go build ./src/bpm
mv ./bpm ./bin/linux/bpm
