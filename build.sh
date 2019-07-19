#!/bin/sh
go install kugg/rules/cmd/game-rules
[ $? -ne 0 ] && exit 1
$GOPATH/src/kugg/rules/test/run-tests.sh
