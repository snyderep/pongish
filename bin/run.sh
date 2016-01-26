#!/bin/bash

set -e

BIN_DIR="$( cd "$( dirname "$0" )" && pwd )"
GO_DIR="$BIN_DIR/../go"

"${GO_DIR}/bin/pongish-darwin-amd64" -config "${BIN_DIR}"/../pongish.config.toml
