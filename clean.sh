#!/bin/sh
set -e
if [ -f env.sh ]
then . ./env.sh
else
    echo 1>&2 "! $0 must be run from the root directory"
    exit 1
fi

xcd() {
    echo
    cd $1
    echo --- cd $1
}

mk() {
    d=$PWD
    xcd $1
    gomake clean
    cd "$d"
}

rm -rf $GOBIN/peg

for pkg in $PKGS
do mk $pkg
done
