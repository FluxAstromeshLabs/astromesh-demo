#!/bin/bash

rm -rf ~/.fluxd
cp -r ./chain/.fluxd ~/.fluxd
DYLD_LIBRARY_PATH=/usr/local/lib ./chain/binary/fluxd start
