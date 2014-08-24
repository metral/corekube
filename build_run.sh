#!/bin/bash

EXPECTEDARGS=2
if [ $# -lt $EXPECTEDARGS ]; then
    echo "Usage: $0 <BRANCH> <NUMBER OF EXPECTED MACHINES>"
    exit 0
fi

DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

BRANCH=$1
MASTER_COUNT=$2
MINION_COUNT=$3
OVERLORD_COUNT=$4

result=`docker build --rm -t setup_kubernetes:$BRANCH $DIR/setup_kubernetes/.`
echo "$result"

echo ""
echo "=========================================================="
echo ""

build_status=`echo $result | grep "Successfully built"`

if [ "$build_status" ] ; then
    docker run -v /tmp:/units -v $DIR/setup_kubernetes/unit_templates:/templates setup_kubernetes:$BRANCH --master_count=$MASTER_COUNT --minion_count=$MINION_COUNT --overlord_count=$OVERLORD_COUNT
fi
