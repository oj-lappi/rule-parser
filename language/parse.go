package language

import (
	"fmt"
	"kugg/compilers/lex"
	"kugg/compilers/parse"
)

//Example Parsetree
//Root
//	ConditionList
//		name (attack)
//		Condition (1)
//			operator (LT)
//			LHS
//				name(targetlevel)
//			RHS
//				name(sourcelevel)
//		Condition (2)
//			operator (GT)
//			LHS
//				name(moverange)
//			RHS
//				name(dist)
//	ConditionList
//		name (move)
//		Condition (1)
//			operator (OR)
//			LHS
//				Condition
//					operator (GT)
//						LHS
//							name (bla)
//						RHS
//							name (ble)
//			RHS
//				Condition
//					operator (GT)
//						LHS
//							name (ble)
//						RHS
//							name (ble)
const (
	ConditionList parse.NodeType = iota
	Condition
	LHS
	RHS
	Ident
	Operator
)

var names = map[parse.NodeType]string{
	ConditionList: "condition-list",
	Condition:     "condition",
	Ident:         "ident",
	Operator:      "op",
	LHS:           "lhs",
	RHS:           "rhs",
}

var arithmeticOperators = map[lex.TokenType]bool{
	EQ: true,
	NE: true,
	GT: true,
	LT: true,
	GE: true,
	LE: true,
}

var booleanOperators = map[lex.TokenType]bool{
	AND: true,
	OR:  true,
}

func init() {
	for k, v := range parse.NodeNames {
		names[k] = v
	}

	parse.NodeNames = names
}

func Parse(source string) (*parse.Tree, error) {
	tree := parse.NewTree("rules", source, parseRoot)
	l := Lex(source)
	err := tree.Parse(l)
	return tree, err
}

func parseRoot(tree *parse.Tree) {
	for {
		tok := tree.Next()
		switch tok.Type() {
		case actionIdentifier:
			if tree.Peek().Type() == leftBrace {
				parseConditionList(tree, tok)
			}
			continue
		case lex.TokenEOF:
			return
		}
		tree.Unexpected(tok, "ident")
	}
}

//parseConditionList parses the rule for the action
func parseConditionList(tree *parse.Tree, nameToken lex.Token) {
	list := tree.AddNonTerminal(ConditionList, nameToken)
	list.AddTerminal(Ident, nameToken)
	tree.Curr = list
	tree.Next() //This is a left brace, and has already been checked by parseRoot
	tree.Next()
	for {
		switch tok := tree.CurrentToken(); tok.Type() {
		case rightBrace:
			list.CommitSubTree()
			tree.Curr = tree.Root
			return
		case name:
			//lhs

			parseCondition(tree, tok)
			tree.Curr = list
		case leftParen:
			//lhs
			parseCondition(tree, tok)
			tree.Curr = list
			//case comment:
			//
		case lex.TokenEOF:
		default:
			tree.ErrorAtTokenf(tok, "unexpected %v, expected a rule", tok)
		}
	}
}

func parseCondition(tree *parse.Tree, firstToken lex.Token) {
	condition := tree.AddNonTerminal(Condition, firstToken)
	tree.Curr = condition.AddNonTerminal(LHS, firstToken)
	parseExpression(tree, firstToken)

	tok := tree.Next()
	if isOperator(tok) {
		condition.AddTerminal(Operator, tok)
	} else {
		tree.ErrorAtTokenf(tok, "%v is not an operator", tok)
	}

	tree.Curr = condition
	parseRHS(tree, tok)
}
func parseRHS(tree *parse.Tree, operatorToken lex.Token) {

	condition := tree.Curr
	rhsToken := tree.Next()

	if typ := rhsToken.Type(); typ == name || typ == leftParen {
		tree.Curr = tree.AddNonTerminal(RHS, rhsToken)

		if isBooleanOperator(operatorToken) {
			parseExpression(tree, rhsToken)
		} else {
			parseIdent(tree, rhsToken)
		}

	} else {
		tree.Unexpected(rhsToken, "( or name")
	}

	next := tree.Next()
	switch typ := next.Type(); {
	case typ == name:
		if tree.NestLevel > 0 {
			tree.ErrorAtTokenf(next, "unclosed (")
			//name -> name isn't valid if inside a parenthesis
		}
		//Return to conditionlist
		return
	case typ == rightParen:
		if tree.NestLevel <= 0 {
			tree.ErrorAtTokenf(next, "unexpected )")
		}
		tree.NestLevel--
		tree.Curr = condition
	case typ == rightBrace:
		if tree.NestLevel > 0 {
			tree.ErrorAtTokenf(next, "unclosed (")
		}
		tree.Curr = condition
		return
	case isArithmeticOperator(next) && isBooleanOperator(operatorToken):
		//Special case, but is it equivalent to the case below?
		//example:
		//a > b && c < d
		//           ^
		//	     We are here
		//the syntax tree should be
		//
		//&&
		//  >
		//    a
		//    b
		//  <
		//    c
		//    d
		//
		//Currently it is
		//
		//&&
		//  >
		//    a
		//    b
		//  c
		//
		//=> c's parent
		//	panic(tree.Curr.Children()[1].Token())
		//tree.Curr == the &&
		//TODO:
		//1. detach c from tree.Curr
		//2. add a new condition, < to tree.Curr
		//3. add a lhs to <
		//4. add the c to the lhs
		//5. add the operator
		//6. recurse
		//	panic(tree.Curr.Children()[0])
		tree.Curr = tree.Curr.Children()[0]
		rotateCondition(tree, next)
		parseRHS(tree, next)

		//The two cases differ only by condition and
	case isBooleanOperator(next):
		//TODO:special case, slurp left side into the lhs of this operation
		//1. detach the already parsed operation from the tree
		//2. add a new Condition
		//2. add a new LHS
		//3. add the operation to the LHS
		//4. continue parsing the operator

		//ALT:
		//1. insert the Condition and LHS ABOVE the current condition and lhs

		tree.Curr = condition
		rotateCondition(tree, next)
		/*
			tree.Curr.RemoveChild(condition)
			tree.Curr = tree.AddNonTerminal(Condition, condition.Token())
			tree.AddNonTerminal(LHS, condition.Token()).AddChild(condition)
			tree.AddTerminal(Operator, next)
		*/
		parseRHS(tree, next)
	default:
		tree.Unexpected(next, "} or condition")
	}
}

func rotateCondition(tree *parse.Tree, newOperator lex.Token) {

	//Remove the old node
	old := tree.Curr
	par := old.Parent()
	fmt.Println(old)
	fmt.Println(par)
	par.RemoveChild(old)
	tree.Curr = par

	//Insert the new node
	tree.Curr = tree.AddNonTerminal(Condition, newOperator)
	tree.AddNonTerminal(LHS, old.Token()).AddChild(old)
	tree.AddTerminal(Operator, newOperator)
}

func parseExpression(tree *parse.Tree, firstToken lex.Token) {
	switch firstToken.Type() {
	case leftParen:
		a := tree.Next()
		b := tree.Next()
		if a.Type() == name && b.Type() == rightParen {
			tree.AddTerminal(Ident, a)
			return
		}
		tree.NestLevel++
		tree.Back()
		parseCondition(tree, a)
	case name:
		tree.AddTerminal(Ident, firstToken)

	default:
		tree.ErrorAtTokenf(firstToken, "unexpected %v, expected an expression", firstToken)
	}
}

//parseIdent is essentially parseExpression except instead of recursing at ( it panics.
func parseIdent(tree *parse.Tree, firstToken lex.Token) {
	switch firstToken.Type() {
	case leftParen:
		a := tree.Next()
		b := tree.Next()
		if a.Type() == name && b.Type() == rightParen {
			tree.AddTerminal(Ident, a)
			return
		}
		tree.ErrorAtTokenf(firstToken, "unexpected %v, expected an identifier", firstToken)
	case name:
		tree.AddTerminal(Ident, firstToken)
	default:
		tree.ErrorAtTokenf(firstToken, "unexpected %v, expected an identifier", firstToken)
	}
}

func isArithmeticOperator(t lex.Token) bool {
	return arithmeticOperators[t.Type()]
}

func isBooleanOperator(t lex.Token) bool {
	return booleanOperators[t.Type()]
}
func isOperator(t lex.Token) bool {
	typ := t.Type()
	return booleanOperators[typ] || arithmeticOperators[typ]
}
