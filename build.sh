#!/bin/sh -ex

BUILDDIR=./build
LATEST_TAG=`git describe --abbrev=0 --tags`

GOOS=linux GOARCH=amd64 go build -o ascendam_${LATEST_TAG}_linux
GOOS=darwin GOARCH=amd64 go build -o ascendam_${LATEST_TAG}_darwin
GOOS=windows GOARCH=amd64 go build -o ascendam_${LATEST_TAG}_windows

