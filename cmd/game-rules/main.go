package main

import (
	"fmt"
	"io/ioutil"
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

//TODO: should return 0 or 1
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
	}

	/*
		parseTree, err := language.Parse(text)
		if err != nil {
			fmt.Println("Errors:")
			fmt.Println(err, "\n")
		}
		fmt.Println("Parse tree:")
		parseTree.PPrint()
	*/
}
