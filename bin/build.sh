#!/bin/bash

set -e

# Cross-compile for several platforms.

BIN_DIR="$( cd "$( dirname "$0" )" && pwd )"
GO_DIR="$BIN_DIR/../go"
APP_NAME="pongish"

# shellcheck disable=SC1090
source "$BIN_DIR/setenv.sh"

function build_arch {
    cd "$GO_DIR" || exit

    echo "> building for $1-$2"

    GOOS=$1 GOARCH=$2 gb build $3
}

function build {
    build_arch "darwin" "amd64" $1
    build_arch "linux" "amd64" $1
    build_arch "windows" "amd64" $1
}

rm -f "$GO_DIR/bin/$APP_NAME*"

# Compile for each platform. Binaries are written to "$GO_DIR/bin".
build "github.com/snyderep/$APP_NAME"

# Binaries are created for each architecture with names like $APP_NAME-darwin-amd64, but an extra
# binary without an extension is created that will run on the platform on which it was built.
# We don't need this extra binary so delete it.
rm -f "$GO_DIR/bin/$APP_NAME"

# Transpile the js.
echo "xpiling"
"$BIN_DIR/xpile.sh"
