#!/bin/bash

# install libs in user local dirs
LIB_PATH=/usr/local/lib
BIN_PATH=/usr/local/bin
ARCH=$(uname -m)
# evmone
echo "installing evmone (evm)"
sudo cp ./chain/libs/evmone/$ARCH/* $LIB_PATH

# golana
echo "installing golana (svm)"
sudo cp ./chain/libs/golana/$ARCH/* $LIB_PATH

# wasmvm
echo "installing lib wasmvm" 
sudo curl -L "https://github.com/CosmWasm/wasmvm/releases/download/v1.5.2/libwasmvm.dylib" -o /usr/local/lib/libwasmvm.dylib

# fluxd
echo "install fluxd"
mkdir -p ./chain/binary
curl -L "https://github.com/FluxAstromeshLabs/astromesh-demo/releases/download/v0.1/fluxd.${ARCH}" -o ./chain/binary/fluxd
chmod +x ./chain/binary/fluxd
sudo cp ./chain/binary/fluxd $BIN_PATH
