#!/bin/bash

DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

result=`docker build --rm -t overlord $DIR/overlord/.`
echo "$result"

echo ""
echo "=========================================================="
echo ""

build_status=`echo $result | grep "Successfully built"`

if [ "$build_status" ] ; then
    docker run -d -v /tmp:/units -v $DIR/overlord/unit_templates:/templates overlord
fi
