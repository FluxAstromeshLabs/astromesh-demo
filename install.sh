#!/bin/bash

# install libs in user local dirs
LIB_PATH=/usr/local/lib
BIN_PATH=/usr/local/bin
ARCH=$(uname -m)
# evmone
if [ ! -f "${LIB_PATH}/libevmone.dylib" ] ; then
    echo "installing evmone (evm)"
    sudo cp ./chain/libs/evmone/$ARCH/* $LIB_PATH
fi

# golana
if [ ! -f "${LIB_PATH}/libgolana.dylib" ] ; then 
    echo "installing golana (svm)"
    sudo cp ./chain/libs/golana/$ARCH/* $LIB_PATH
fi

# wasmvm
if [ ! -f "${LIB_PATH}/libwasmvm.dylib" ] ; then
    echo "installing lib wasmvm" 
    sudo curl -L "https://github.com/CosmWasm/wasmvm/releases/download/v1.5.2/libwasmvm.dylib" -o /usr/local/lib/libwasmvm.dylib
fi

# fluxd
if [ ! -f "${BIN_PATH}/fluxd" ] ; then 
    mkdir -p ./chain/binary
    curl -L "https://github.com/FluxAstromeshLabs/astromesh-demo/releases/download/v0.1/fluxd.${ARCH}" -o ./chain/binary/fluxd
    chmod +x ./chain/binary/fluxd
    sudo cp ./chain/binary/fluxd $BIN_PATH
fi
