#!/bin/bash

set -Eeuo pipefail

(cd grammars/c/ && go generate)
(cd grammars/calculator/ && go generate)
(cd grammars/calculatorast/ && go generate)
(cd grammars/fexl/ && go generate)
(cd grammars/java/ && go generate)
(cd grammars/longtest/ && go generate)