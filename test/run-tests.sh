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

runopt=`getopt -n $0 -o h?Vvs -l help,verbose,silent,Verbose, -- "$@"`
if [ $? != 0 ] ; then echo "Exiting" >&2 ; exit 1 ; fi
eval set -- "$runopt"

silent_all=1
silent_errors=1
successes=0
attempts=0

while true;do
case "$1" in

        -h|--help|?)
                print_usage
                exit 0
                ;;
	-v|--verbose)
		silent_errors=0
		;;
	-V|--Verbose)
		silent_all=0
		silent_errors=0
		;;
	-s|--silent)
		;;
        --)
                shift
                break
                ;;
        * ) break ;;

esac
shift
done

errors=$(mktemp errors-testing.XXXXXXX)
trap "rm -f -- $errors" EXIT

verbose_output(){
	echo "$line"
	echo "$output"
        echo ""
}

test_file() {
	file=$1
	expect=$2
	fail=0
	let attempts++
	if [ $expect == 0 ];then
		expecting="no errors:"
	else
		expecting="an error: "
	fi
	printf "%s %s %s"  "Testing $file expecting" "$expecting" "${padding:${#file}}"
	output=$(game-rules $testdir/$file 2>$errors)
	if [ $? != $expect ];then
		fail=1
		echoF FAIL
		echoF "$(<$errors)"
		echo ""
	else
		let successes++
		echoS PASS
	fi

	[ $silent_all -eq 0 ] && verbose_output && return
	[ $silent_errors -eq 0 ] && [ $fail -eq 1 ] && verbose_output
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
