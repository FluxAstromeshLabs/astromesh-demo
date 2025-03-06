#!/bin/bash

# install libs in user local dirs
LIB_PATH=/usr/local/lib
BIN_PATH=/usr/local/bin
ARCH=$(uname -m)
# install evmone
if [ ! -f "${LIB_PATH}/libevmone.dylib" ] ; then
    sudo cp ./chain/libs/evmone/$ARCH/* $LIB_PATH
fi

# install golana
if [ ! -f "${LIB_PATH}/libgolana.dylib" ] ; then 
    sudo cp ./chain/libs/golana/$ARCH/* $LIB_PATH
fi

# fluxd
if [ ! -f "${BIN_PATH}/fluxd" ] ; then 
    sudo cp ./chain/binary/$ARCH/fluxd $BIN_PATH
fi
