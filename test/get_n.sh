#!/bin/bash
curl 127.0.0.1:8123/stats
for id in $(seq $1 $2)
do
  curl 127.0.0.1:8123/hash/$id 
done

