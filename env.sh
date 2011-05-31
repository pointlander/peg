if test -z "$GOROOT"
then
    # figure out what GOROOT is supposed to be
    GOROOT=`printf 't:;@echo $(GOROOT)\n' | gomake -f -`
    export GOROOT
fi

PKGS="
    cmd/bootstrap
    cmd/peg
    pkg/calculator
    pkg/stl
    cmd/calculator
"
TESTS="
    pkg/calculator
    pkg/stl
"