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
	AND                            //&&
	OR                             //||
	name                           //alphabetic word
	value                          //numeric value
	comment                        //#
)

func Lex(source string) *lex.BaseLexer {
	lexer := lex.Lex("rules", source)
	lexer.Run(lexRoot)
	return lexer
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
		//TODO:comments have only been implemented outside blocks...
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
	return nil
}

//lexAction lexes a top level identifier identifying the action to which this
//applies
func lexActionIdentifier(lexer *lex.BaseLexer) lex.StateFn {

	//Valid identifier runes are _- and [:alnum:]
	lexer.AcceptMatchOrRangeRun("_-", unicode.Letter, unicode.Digit)
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
		return lexLHS
	}

	lexer.NotAllowedInContext(r, "action identifier")
	return nil

}

func lexLHS(lexer *lex.BaseLexer) lex.StateFn {

	lexer.IgnoreSpaces()

	//Looking for ( or ident

	r := lexer.Next()

	switch {
	case unicode.IsLetter(r):
		return lexLHSOperand
	case r == '(':
		lexer.Emit(leftParen)
		return lexLHS
	case r == ')':
		//TODO: this is the only iffy part...
		//But I believe I can now catch all VALID sentences, while allowing some false positives from the lexer
		//We will catch those in the parser
		lexer.Emit(rightParen)
		return lexLHS
	case r == '}':
		lexer.Emit(rightBrace)
		return lexActionIdentifier
	}

	lexer.Unexpected(r, "identifier or left paren")
	return nil
}

//name
func lexLHSOperand(lexer *lex.BaseLexer) lex.StateFn {
	lexer.AcceptMatchOrRangeRun("_-", unicode.Letter, unicode.Digit)
	lexer.Emit(name)

	lexer.IgnoreSpaces()

	return lexOperator

}

//The leftmost rune has NOT been lexed
func lexOperator(lexer *lex.BaseLexer) lex.StateFn {
	r := lexer.Next()
	switch r {
	case '=':
		return lexEquals
	case '!':
		return lexInequals
	case '>':
		return lexGreater
	case '<':
		return lexLess
	case '&':
		return lexAnd
	case '|':
		return lexOr
	}
	lexer.Unexpected(r, "operator")
	return nil
}

func lexEquals(lexer *lex.BaseLexer) lex.StateFn {
	second := lexer.Next()
	if second != '=' {
		lexer.Back()
	}
	lexer.Emit(EQ)
	return lexRHS
}

func lexInequals(lexer *lex.BaseLexer) lex.StateFn {
	second := lexer.Next()
	if second != '=' {
		lexer.Back()
		lexer.Unexpected(second, "=")
	}
	lexer.Emit(NE)
	return lexRHS
}

func lexGreater(lexer *lex.BaseLexer) lex.StateFn {
	second := lexer.Next()
	if second != '=' {
		lexer.Back()
		lexer.Emit(GT)
		return lexRHS
	}
	lexer.Emit(GE)
	return lexRHS

}

func lexLess(lexer *lex.BaseLexer) lex.StateFn {
	second := lexer.Next()
	if second != '=' {
		lexer.Back()
		lexer.Emit(LT)
		return lexRHS
	}
	lexer.Emit(LE)
	return lexRHS
}

func lexAnd(lexer *lex.BaseLexer) lex.StateFn {
	second := lexer.Next()
	if second != '&' {
		lexer.Back()
	}
	lexer.Emit(AND)
	return lexRHS
}

func lexOr(lexer *lex.BaseLexer) lex.StateFn {
	second := lexer.Next()
	if second != '|' {
		lexer.Back()
	}
	lexer.Emit(OR)
	return lexRHS
}

//have just lexed an operator
func lexRHS(lexer *lex.BaseLexer) lex.StateFn {
	lexer.IgnoreSpaces()

	r := lexer.Next()
	switch {
	case unicode.IsLetter(r):
		return lexRHSOperand
	case r == '(':
		lexer.Emit(leftParen)
		return lexLHS
	}
	lexer.Unexpected(r, "identifier or left paren")
	return nil
}

func lexRHSOperand(lexer *lex.BaseLexer) lex.StateFn {
	lexer.AcceptMatchOrRangeRun("_-", unicode.Letter, unicode.Digit)
	lexer.Emit(name)

	lexer.IgnoreSpaces()

	return lexLHS
}
