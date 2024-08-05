#!/bin/bash

set -e
set -x

cd ~/go-api

systemctl --user stop go-api.service

git pull -v

time go test -test.v ./... && time go build -x

sudo setcap cap_net_bind_service=+ep ./go-api

systemctl --user restart go-api.service
