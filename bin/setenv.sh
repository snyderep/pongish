#!/bin/bash

# source this file

BIN_DIR="$( cd "$( dirname "$0" )" && pwd )"
GO_DIR="$BIN_DIR/../go"

pathadd() {
    if [ -d "$1" ] && [[ ":$PATH:" != *":$1:"* ]]; then
        PATH="${PATH:+"$PATH:"}$1"
    fi
}

pathadd "$GO_DIR/boot/bin"

export PATH="$PATH"
