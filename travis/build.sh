#!/bin/bash -e

git clone https://github.com/metral/corekube_travis
pushd corekube_travis/infra_test
./setup.sh
