#!/bin/sh

if [ -f /etc/timezone ]; then
	TZ=$(cat /etc/timezone)
elif [ -f /etc/localtime ]; then
	TZ=$(ls -la /etc/localtime | cut -d/ -f8-9)
else
	TZ=UTC
fi

docker rm -f mlbme >/dev/null 2>&1

docker run \
	--env TZ=${TZ} \
    --dns=1.1.1.1 \
	--name mlbme \
	-p 6789:6789 \
	-it dtpoole/mlbme $@