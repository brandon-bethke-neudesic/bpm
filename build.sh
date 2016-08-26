#!/bin/bash
export GOPATH=`pwd`
go get bpm
go install bpm
if [ -f /usr/local/bin/bpm ]; then
    rm /usr/local/bin/bpm
fi
echo Creating symbolic link for bpm in /usr/local/bin
ln -s `pwd`/bin/bpm /usr/local/bin/bpm

echo Creating linux version
env GOPATH=`pwd` GOOS=linux GOARCH=amd64 go build ./src/bpm
mkdir -p ./bin/linux
mv ./bpm ./bin/linux/bpm

echo Creating darwin version
env GOPATH=`pwd` GOOS=darwin GOARCH=amd64 go build ./src/bpm
mkdir -p ./bin/darwin
mv ./bpm ./bin/darwin/bpm
