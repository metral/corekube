#!/bin/bash

EXPECTEDARGS=2
if [ $# -lt $EXPECTEDARGS ]; then
    echo "Usage: $0 <BRANCH> <NUMBER OF EXPECTED MACHINES>"
    exit 0
fi

DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PARENTDIR="$(dirname $DIR)"

BRANCH=$1
MACHINE_COUNT=$2

result=`docker build --rm -t setup:$BRANCH $PARENTDIR/setup/.`
echo "$result"

echo ""
echo "=========================================================="
echo ""

build_status=`echo $result | grep "Successfully built"`

if [ "$build_status" ] ; then
    docker run setup:$BRANCH --machine_count=$MACHINE_COUNT
fi
