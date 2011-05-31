#!/bin/sh
set -e

if ! which roundup > /dev/null
then
    echo "You need to install roundup to run this script."
    echo "See: http://bmizerany.github.com/roundup"
    exit 1
fi

if [ -f env.sh ]
then . ./env.sh
else
    echo 1>&2 "! $0 must be run from the root directory"
    exit 1
fi

{
    for pkg in $TESTS
    do
        name=$(echo $pkg | sed 's/\//_/' | tr -d .)
        echo "it_passes_$name() { cd $pkg; gotest $@; }"
    done
} > all-test.sh

: ${GOMAXPROCS:=10}
export GOMAXPROCS

trap 'rm -f all-test.sh' EXIT INT
roundup
