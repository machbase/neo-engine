#!/bin/sh

OS=`uname -o | tr  '[:upper:]' '[:lower:]'`
ARCH=`uname -m`

gcc -lmachengine_standard_${OS}_${ARCH} \
    -L../../native \
    -I../../native \
    -o machcli main.c && \
./machcli