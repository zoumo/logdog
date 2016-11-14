#!/bin/bash

set -e


command_exists() {
    command -v "$@" > /dev/null 2>&1
}

if ! command_exists godep; then
    go get github.com/tools/godep
fi

godep get

go test .