#!/bin/bash

print_usage(){
cat <<EOF
Usage: run_tests.sh [--silent|--verbose]
Will run tests for game-rules on a set of example texts

Verbose mode will print the output of the lexer and parser
Silent mode will redirect to /dev/null

Silent mode is the default
EOF
}

runopt=`getopt -n $0 -o h?vs -l help,verbose,silent, -- "$@"`
if [ $? != 0 ] ; then echo "Exiting" >&2 ; exit 1 ; fi
eval set -- "$runopt"

silent=1
successes=0
attempts=0

while true;do
case "$1" in

        -h|--help|?)
                print_usage
                exit 0
                ;;
	-v|--verbose)
		silent=0
		;;
	-s|--silent)
		silent=1
		;;
        --)
                shift
                break
                ;;
        * ) break ;;

esac
shift
done

test_file() {
	file=$1
	expect=$2
	let attempts++
	if [ $expect != 0 ];then
		expecting="no errors:"
	else
		expecting="an error: "
	fi
	printf "%s %s %s"  "Testing $file expecting" "$expecting" "${padding:${#file}}"
	output=$(game-rules $testdir/$file)
	if [ $? != $expect ];then
		echoF FAIL
	else
		let successes++
		echoS PASS
	fi
	if [ $silent -ne 1 ];then
		echo "$line"
		echo "$output"
		echo ""
        fi
}

succeed() {
	test_file $1 0
}

fail() {
	test_file $1 1
}

echoS(){
	colorFmt "\e[32m" "$1"
}

echoF(){
	colorFmt "\e[31m" "$1"
}

colorFmt(){
	printf "%b%s%b\n" "$1" "$2" "\e[0m"
}

testdir=$GOPATH/src/kugg/rules/test

     #Testing
line="==OUT=="
padding="              "
succeed empty-block    #
succeed simple-rule    #
succeed paren-lhs      #
succeed paren-rhs      #
succeed multiple-rules #

if [ $successes == $attempts ];then
	echoS "$successes/$attempts successful tests"
	echoS "Tests PASSED"
	exit 0
else
	echoF "$successes/$attempts successful tests"
	echoF "Tests FAILED"
	exit 1
fi
