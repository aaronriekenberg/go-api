#!/usr/pkg/bin/bash

cd ~/go-api

pkill go-api

nohup ./go-api ./configfiles/rpi.toml 2>&1 | ~/bin/simplerotate logs &
