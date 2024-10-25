#!/bin/bash

export GOGC=1000
export GOMEMLIMIT=1GiB

echo "GO variables"
env | grep GO

exec ./go-api ./configfiles/$(hostname)-config.toml
