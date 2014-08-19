#!/bin/bash

EXPECTEDARGS=2
if [ $# -lt $EXPECTEDARGS ]; then
    echo "Usage: $0 <DOCKER_REPO_TAG> <NUMBER OF EXPECTED MACHINES>"
    exit 0
fi

DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PARENTDIR="$(dirname $DIR)"

BRANCH=$1
MACHINE_COUNT=$2

if [ "$BRANCH" != "latest" ]; then
    git checkout -b $BRANCH origin/$BRANCH
fi

result=`docker build --rm -t etcd_nodes $PARENTDIR/etcd_nodes/.`
echo "$result"

echo ""
echo "=========================================================="
echo ""

build_status=`echo $result | grep "Successfully built"`

if [ "$build_status" ] ; then
    docker run -e DOCKERHOST_HOSTNAME=`hostname` etcd_nodes --machine_count=$MACHINE_COUNT
fi
