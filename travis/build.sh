#!/bin/bash -e

go get github.com/tools/godep

git clone https://github.com/metral/corekube_travis
pushd corekube_travis/corekube_test
echo "========================================"
echo "corekube_travis commit: `git rev-parse --short HEAD`"
echo "========================================"
godep get ./...
popd

# Copy conf.json from overlord in godeps to
# /tmp where overlord's lib expects it - we use lib in the corekube_test to
# piece together the etcd api & client port
mkdir -p /tmp/
cp $HOME/gopath/src/github.com/metral/overlord/conf.json /tmp/
