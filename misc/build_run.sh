#!/bin/bash

result=`docker build --rm -t etcd_nodes ../etcd_nodes/.`
echo "$result"

echo ""
echo "=========================================================="
echo ""

build_status=`echo $result | grep "Successfully built"`

if [ "$build_status" ] ; then
    docker run etcd_nodes
fi
