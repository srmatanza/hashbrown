#!/bin/bash
curl 127.0.0.1:8123/stats
for pw in $(seq 1 $1)
do
  curl -X POST 127.0.0.1:8123/hash --data "password=abc$pw"
done

for id in $(seq 1 $1)
do
  curl 127.0.0.1:8123/hash/$id 
done

#sleep 5
curl -X POST 127.0.0.1:8123/hash --data "password=lastpass"
curl -X POST 127.0.0.1:8123/shutdown

for id in $(seq 1 $1)
do
  curl 127.0.0.1:8123/hash/$id
done

curl 127.0.0.1:8123/stats

