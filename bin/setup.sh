#!/bin/bash

# This project uses the gb tool for build and therefore doesn't really care about GOPATH,
# but the gb tool itself is fetched via "go get", so to update/install gb we set GOPATH.


BIN_DIR="$( cd "$( dirname "$0" )" && pwd )"
GO_DIR="$BIN_DIR/../go"

# shellcheck disable=SC1090
source "$BIN_DIR/setenv.sh"

export GOPATH="$GO_DIR/boot:$GO_DIR:$GO_DIR/vendor"

# gb build tool, see http://getgb.io
go get -u github.com/constabulary/gb/...

unset GOPATH
