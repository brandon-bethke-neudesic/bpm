#!/bin/bash
go get bpm
go install bpm
if [ -f /usr/local/bin/bpm ]; then
    rm /usr/local/bin/bpm
fi
ln -s `pwd`/bin/bpm /usr/local/bin/bpm

env GOPATH=`pwd` GOOS=linux GOARCH=amd64 go build ./src/bpm
mv ./bpm ./bin/linux/bpm
