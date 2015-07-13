#!/bin/bash

realpath() {
    [[ $1 = /* ]] && echo "$1" || echo "$PWD/${1#./}"
}

REALPATH=`realpath $0`
REALPATH=`dirname $REALPATH`
if [ -z .ibenchmark/src/github.com/albus01/ibenchmark ];then
    rm -rf .ibenchmark/src/github.com/albus01/ibenchmark/*
else
    mkdir -p '.ibenchmark/src/github.com/albus01/ibenchmark'
fi
export GOPATH="$REALPATH/.ibenchmark"

echo "GOPATH=$GOPATH"

cp -r {gospdy,bench} .ibenchmark/src/github.com/albus01/ibenchmark/

echo
echo 'You can now use `make` to rebuild'
echo

make
