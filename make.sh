#!/bin/bash
CGO_ENABLED=0

# Install dependencies
# Runtime dependencies
go get github.com/tools/godep



BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
VERSION=`cat version.go | grep "const Version string" | cut -d'"' -f2`

# Build for OSX
godep go build -ldflags "-w -X main.buildDate ${BUILD_DATE} -X main.version ${VERSION}" -o builds/autoscale-grapher-$VERSION.osx
if [ $? -eq 0 ]; then
  echo "Success Build artifact - builds/tropo-grapher-$VERSION.osx"
else
  echo "Build error"
  exit $?
fi

# Build for Linux
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build -ldflags "-w -X main.buildDate ${BUILD_DATE} -X main.version ${VERSION}" -o builds/tropo-grapher-$VERSION.linux
if [ $? -eq 0 ]; then
  echo "Success Build artifact - builds/tropo-grapher-$VERSION.linux"
else
  echo "Build error"
  exit $?
fi
