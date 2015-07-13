#!/bin/bash

realpath() {
    [[ $1 = /* ]] && echo "$1" || echo "$PWD/${1#./}"
}

REALPATH=`realpath $0`
REALPATH=`dirname $REALPATH`

mkdir -p '.ibenchmark/src/github.com/albus01/ibenchmark'
export GOPATH="$REALPATH/.ibenchmark"

echo "GOPATH=$GOPATH"

cp -r {gospdy,bench} .ibenchmark/src/github.com/albus01/ibenchmark/

if [ $? != 0 ]; then
    echo
    echo 'You need the `go` command line tool'
    echo "More information at http://www.golang.org"
    echo

    # ubuntu: apt-get install golang golang-go

    # mac os: http://www.golang.org/

    exit 1
fi

echo
echo 'You can now use `make` to rebuild'
echo

make
