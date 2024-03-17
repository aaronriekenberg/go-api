#!/usr/pkg/bin/bash

cd /home/aaron/go-api

pkill go-api

nohup ./go-api ./configfiles/rpi.toml > output 2>&1 &
