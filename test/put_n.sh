#!/bin/bash
curl 127.0.0.1:8123/stats
for pw in $(seq 1 $1)
do
  curl -s -X POST 127.0.0.1:8123/hash --data "password=abc$pw" > /dev/null &
done

curl 127.0.0.1:8123/stats

