#!/bin/bash

if [ $# -ne 3 ];
then
    echo "usage: update.sh <edition> <os> <arch> (ex: update.sh fog linux amd64)"
    exit 1
fi

echo "cp $MACHBASEDEV_HOME/mm/src/include/machEngine.h ./machEngine.h"
cp $MACHBASEDEV_HOME/mm/src/include/machEngine.h ./machEngine.h
echo "cp $MACHBASEDEV_HOME/machbase_home/lib/libmachengine.a ./libmachengine_$1_$2_$3.a"
cp $MACHBASEDEV_HOME/machbase_home/lib/libmachengine.a ./libmachengine_$1_$2_$3.a
