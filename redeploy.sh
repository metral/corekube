#!/bin/bash

set -e

if docker ps | grep -q overlord; then
  docker rm -f overlord
fi

function seen_exists() {
  if etcdctl ls | grep --quiet /seen; then
    return 0
  else
    return 1
  fi
}

if seen_exists; then
  etcdctl rm seen
  sleep 1
  if seen_exists; then
    echo "etcdctl \"seen\" keys still exist"
    exit 1
  fi
fi

fleetctl list-units | awk '{print $1}' | grep -v UNIT | xargs fleetctl destroy
sleep 1

if [ $(fleetctl list-units | wc -l) -gt 1 ]; then
  echo "Some fleet units still exist"
  exit 1
fi

pushd /root/overlord
echo "Building overlord image, this could take a while..."
./build_run.sh >/dev/null
popd

echo "Complete! Monitor overlord status via \"docker logs overlord\" and wait until its complete"
