#!/bin/bash

EXPECTEDARGS=1
if [ $# -lt $EXPECTEDARGS ]; then
    echo "Usage: $0 <BRANCH>"
    exit 0
fi

DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

BRANCH=$1

result=`docker build --rm -t overlord:$BRANCH $DIR/overlord/.`
echo "$result"

echo ""
echo "=========================================================="
echo ""

build_status=`echo $result | grep "Successfully built"`

if [ "$build_status" ] ; then
    docker run -v /tmp:/units -v $DIR/overlord/unit_templates:/templates overlord:$BRANCH
fi
