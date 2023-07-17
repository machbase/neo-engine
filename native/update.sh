#!/bin/bash

if [ $# -ne 2 ];
then
    echo "usage: update.sh <os> <arch> (ex: update.sh linux amd64)"
    exit 1
fi

echo "cp $MACHBASEDEV_HOME/mm/src/include/machEngine.h ./machEngine.h"
cp $MACHBASEDEV_HOME/mm/src/include/machEngine.h ./machEngine.h
echo "cp $MACHBASEDEV_HOME/machbase_home/lib/libmachengine.a ./libmachengine_standard_$1_$2.a"
cp $MACHBASEDEV_HOME/machbase_home/lib/libmachengine.a ./libmachengine_standard_$1_$2.a
