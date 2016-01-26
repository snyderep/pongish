#!/bin/bash

set -e

BIN_DIR="$( cd "$( dirname "$0" )" && pwd )"
GO_DIR="$BIN_DIR/../go"
APP_NAME="pongishweb"

# If the gopherjs command is missing then create it.
if [ ! -f "$GO_DIR/bin/gopherjs" ];
then
   echo "building gopherjs"
   cd "$GO_DIR"
   gb build "github.com/gopherjs/gopherjs"
fi

# Transpile the gopherjs
# gopherjs works like the regular go command and requires a GOPATH.
rm -f $GO_DIR/bin/*.js
rm -f $GO_DIR/bin/*.js.map
export GOPATH="$GO_DIR:$GO_DIR/vendor"
# This will create bin/ $APP_NAME.js and $APP_NAME.js.map.
# "$GO_DIR/bin/gopherjs" -m install "github.com/snyderep/$APP_NAME"
"$GO_DIR/bin/gopherjs" install "github.com/snyderep/$APP_NAME"
unset GOPATH

# Copy the transpiled js to the static dir
GOLANG_JS="$BIN_DIR/../static/golangjs"
rm -f "$GOLANG_JS/*.js"
rm -f "$GOLANG_JS/*.js.map"
cp -f "$GO_DIR/bin/"*.js "$BIN_DIR/../static/golangjs"
cp -f "$GO_DIR/bin/"*.js.map "$BIN_DIR/../static/golangjs"
