#!/bin/bash -e

go get github.com/tools/godep

git clone https://github.com/metral/corekube_travis
pushd corekube_travis
godep get ./...
popd
