#!/bin/bash

./install.sh

rm -rf ~/.fluxd
cp -r ./chain/.fluxd ~/.fluxd
DYLD_LIBRARY_PATH=/usr/local/lib fluxd start
