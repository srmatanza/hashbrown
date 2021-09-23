#!/bin/bash
curl 127.0.0.1:8123/stats
curl -X POST 127.0.0.1:8123/hash --data "password=abc"
curl -X POST 127.0.0.1:8123/hash --data "password=123"
curl -X POST 127.0.0.1:8123/hash --data "password=password"

curl 127.0.0.1:8123/hash/1
curl 127.0.0.1:8123/hash/2
curl 127.0.0.1:8123/hash/3
sleep 5

curl 127.0.0.1:8123/hash/1
curl 127.0.0.1:8123/hash/2
curl 127.0.0.1:8123/hash/3

curl 127.0.0.1:8123/stats

