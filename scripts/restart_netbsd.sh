#!/usr/pkg/bin/bash

cd /home/aaron/go-api

pkill go-api

nohup ./go-api ./configfiles/rpi.toml 2>&1 | simplerotate logs &
