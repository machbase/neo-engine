#!/bin/bash

set -e
DIR=$(dirname "${BASH_SOURCE[0]}")

$DIR/machsvr --pname machsvr --pid $DIR/machsvr.pid -c $DIR/conf/machsvr.hcl