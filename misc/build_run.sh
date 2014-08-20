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
    docker run -e DOCKERHOST_HOSTNAME=`hostname` -v /etc:/host_etc setup:$BRANCH --machine_count=$MACHINE_COUNT
fi

/usr/bin/systemctl restart fleet
/usr/bin/sleep 3
/usr/bin/fleetctl list-machines
