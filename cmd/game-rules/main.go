package main

import (
	"fmt"
	"io/ioutil"
	"kugg/compilers/lex"
	"kugg/rules/language"
	"log"
	"os"
)

const name = "game-rules"

const usage = `
%[1]s is a front end for the %[1]s DSL.
You can use it to test the validity of a %[1]s configuration.

Usage:
%[1]s path_to_source 

Prints out errors and the parse tree of the source
`

func main() {
	var fileName string
	fileName = os.Args[1]
	switch fileName {
	case "", "?", "h", "-h":
		fmt.Printf(usage, name)
		return
	}
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	text := string(b)

	lexer := language.Lex(text)
	for t := range lexer.Tokens {
		fmt.Println(t)
		if t.Type() == lex.TokenError {
			fmt.Fprintf(os.Stderr, "error: %v\n", t)
			os.Exit(1)
		}
	}

	parseTree, err := language.Parse(text)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Parse tree:")
	parseTree.PPrint()

	fmt.Fprintf(os.Stderr, "error: %v\n", "TEST")
	os.Exit(1)
}
