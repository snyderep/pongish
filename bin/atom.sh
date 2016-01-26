#!/bin/bash

# The atom editor package go-plus does not yet support the gb build tool. The package is not aware that
# gb puts vendored sources under a vendor directory. In addition the package adds its own go dependencies
# to the first path in GOPATH. We don't want any atom go-plus dependencies polluting our source.

# This script assumes that atom is installed and the atom command is on the PATH.

BIN_DIR="$( cd "$( dirname "$0" )" && pwd )"
GO_DIR="$BIN_DIR/../go"

# Build a GOPATH so the atom go-plus plugin is happy. Put a 'global' path first so go-plus installs its
# junk somewhere other than in our project.
GOPATH="$GO_DIR:$GO_DIR/vendor" atom .
