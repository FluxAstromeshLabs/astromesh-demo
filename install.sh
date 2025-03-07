#!/bin/bash

# install libs in user local dirs
LIB_PATH=/usr/local/lib
BIN_PATH=/usr/local/bin
ARCH=$(uname -m)
# evmone
if [ ! -f "${LIB_PATH}/libevmone.dylib" ] ; then
    sudo cp ./chain/libs/evmone/$ARCH/* $LIB_PATH
fi

# golana
if [ ! -f "${LIB_PATH}/libgolana.dylib" ] ; then 
    sudo cp ./chain/libs/golana/$ARCH/* $LIB_PATH
fi

# fluxd
if [ ! -f "${BIN_PATH}/fluxd" ] ; then 
    mkdir -p ./chain/binary
    curl -L "https://github.com/FluxAstromeshLabs/astromesh-demo/releases/download/v0.1/fluxd.${ARCH}" -o ./chain/binary/fluxd
    chmod +x ./chain/binary/fluxd
    sudo cp ./chain/binary/fluxd $BIN_PATH
fi
