#!/bin/bash

env GOOS=linux GOARCH=amd64 go build

chmod +x nightlight-router

mv nightlight-router /Users/martezreed/custom-linux-os/binaries