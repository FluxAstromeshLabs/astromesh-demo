#!/bin/bash

./install.sh

rm -rf ~/.fluxd
cp -r ./chain/.fluxd ~/.fluxd
fluxd start
