#!/bin/bash -e

go get github.com/tools/godep

git clone https://github.com/metral/corekube-travis
pushd corekube-travis
godep get ./...
ls -alh $GOPATH/src/github.com
ls -alh $GOPATH/bin
popd
