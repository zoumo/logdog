#!/bin/bash

set -e

cd "$(dirname "$BASH_SOURCE")/.."

command_exists() {
    command -v "$@" > /dev/null 2>&1
}

if ! command_exists godep; then
    go get https://github.com/tools/godep.git
fi

godep get

go test .