#!/bin/bash -e

go get github.com/tools/godep

git clone https://github.com/metral/corekube-travis
pushd corekube-travis
godep get ./...
ls -alh $HOME/gopath/
popd
