package language

import (
	"kugg/compilers/parse"
)

func Parse(source string) (error, *parse.Tree) {
	tree := parse.NewTree("rules", source, parseRoot)
	tree.Parse(lex(source))
}
