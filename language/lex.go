package language

import (
	"kugg/compilers/lex"
	"unicode"
)

const (
	actionIdentifier lex.TokenType = iota
	leftBrace                      //{
	rightBrace                     //}
	leftParen                      //(
	rightParen                     //)
	EQ                             //== (or =)
	NE                             //!=
	GT                             //>
	LT                             // <
	GE                             // >=
	LE                             //<=
	name                           //alphabetic word
	value                          //numeric value
	comment                        //#
)

func Lex(source string) {
	lexer := lex.NewLexer("rules", source)
	lexer.Run(lexRoot)
}

func lexRoot(lexer *lex.BaseLexer) lex.StateFn {
top:
	lexer.IgnoreSpaces()

	letterFound := lexer.AcceptUnicodeRanges(unicode.Letter)
	if letterFound {
		return lexActionIdentifier
	}
	r := lexer.Next()
	switch r {
	case lex.EOF:
		lexer.Emit(lex.TokenEOF)
		return nil
	case '#':
		lexer.AcceptUntil("\n")
		//If we want to use comments for e.g. automatic
		//generation of manuals/instructions
		//lexer.Emit(comment)
		lexer.Next()
		lexer.Ignore()
		goto top
	}
	lexer.Unexpected(r, "identifier")

}

//lexAction lexes a top level identifier identifying the action to which this
//applies
func lexActionIdentifier(lexer *lex.BaseLexer) lex.StateFn {

	//Valid identifier runes are _- and [:alnum:]
	lexer.AcceptMatchOrRangeRun("_-", unicode.Letters, unicode.Digits)
	lexer.Emit(actionIdentifier)

	lexer.IgnoreSpaces()

	r := lexer.Next()
	switch r {
	case lex.EOF:
		//This is an empty rule, no conditions
		lexer.Emit(lex.TokenEOF)
		return nil
	case '{':
		lexer.Emit(leftBrace)
		return lexInsideAction
	}

	lexer.NotAllowedInContext(r, "action identifier")
	return nil

}

func lexInsideAction(lexer *lex.BaseLexer) lex.StateFn {
	//Looking for ( or ident

	lexer.IgnoreSpaces()
	letterFound := lexer.AcceptUnicodeRanges(unicode.Letter)
	if letterFound {
		lexer.AcceptMatchOrRangeRun("_-", unicode.Letters, unicode.Digits)
		lexer.Emit(name)
		return lexOperator
	}
	r := lexer.Next()
	switch r {
	case '(':
		lexer.Emit(leftParen)
		//TODO:you stopped here
	}
}
