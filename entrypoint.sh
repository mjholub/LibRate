#!/bin/sh

set -e

air -c /app/data/.air.toml

while true; do
	if fswatch -r /app/data/air.log; then
		# send logs to promethes
		curl --data-binary @/app/data/air.log http://prometheus:9091/metrics
	fi
	sleep 2
done
