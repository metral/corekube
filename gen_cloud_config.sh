#!/bin/bash

#TOKEN=$(curl -s https://discovery.etcd.io/new)
TOKEN="http://$IP:4001/v2/keys/testcluster"
CLOUD_CONFIG=$(sed -e "s#<discovery_token>#$TOKEN#g" data.yaml)

heat --os-password $HEAT_OS_PASSWORD stack-create Test \
    -f heat-stable.yaml \
    -P key-name=argon_iad \
    -P flavor='2 GB Performance' \
    -P count=3 \
    -P user-data="$CLOUD_CONFIG" \
    -P name="CoreOS-stable"
