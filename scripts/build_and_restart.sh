#!/bin/bash

set -e
set -x

cd ~/go-api

systemctl --user stop go-api.service

git pull -v

time go build -x

systemctl --user restart go-api.service
