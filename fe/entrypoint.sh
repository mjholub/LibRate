#!/bin/sh

set -e

while true; do
	if fswatch -r fe; then
		if pnpm run build; then
			# refer to docs for how to set up prometheus to receive notifications
			# on build statuses via notify-send
			echo "Frontend Build Successful"
			# prometheus container must be set up with this hostname, on librate network
			echo 'frontend_build_status{status="success"} 1' | curl --data-binary @- http://prometheus:9091/metrics
		else
			echo "Librate frontend build failed"
			echo 'frontend_build_status{status="failure"} 1' | curl --data-binary @- http://prometheus:9091/metrics
		fi
	fi
	sleep 10
done
