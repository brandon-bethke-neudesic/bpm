#!/bin/bash
if [[ "$1" != "preserve-go-path" ]]; then
    export GOPATH=`pwd`
fi

go get bpm
if [[ $? -ne 0 ]]; then
    exit 1
fi
go install bpm
if [[ $? -ne 0 ]]; then
    exit 1
fi
if [ -f /usr/local/bin/bpm ]; then
    echo Removing existing symbolic link
    rm /usr/local/bin/bpm
fi
echo Creating symbolic link for bpm in /usr/local/bin
ln -s $GOPATH/bin/bpm /usr/local/bin/bpm

echo Creating linux version
mkdir -p $GOPATH/bin/linux
env GOPATH=$GOPATH GOOS=linux GOARCH=amd64 go build -o $GOPATH/bin/linux/bpm bpm

if [[ "$TPM_CLUSTER_HOME" != "" ]]; then
	echo Copying linux binary to tpm cluster includes location
	cp $GOPATH/bin/linux/bpm $TPM_CLUSTER_HOME/includes/bpm/bin/linux
fi

#echo Creating darwin version
#mkdir -p $GOPATH/bin/darwin
#env GOPATH=$GOPATH GOOS=darwin GOARCH=amd64 go build -o $GOPATH/bin/darwin/bpm bpm
