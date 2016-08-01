#!/bin/bash
go get bpm
go install bpm
if [ -f /usr/local/bin/bpm ]; then
    rm /usr/local/bin/bpm
fi
ln -s `pwd`/bin/bpm /usr/local/bin/bpm
