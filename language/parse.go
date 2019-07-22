package language

import (
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
	LeftOperand
	RightOperand
	ident
	operator
)

var names = map[parse.NodeType]string{
	ConditionList: "condition-list",
	Condition:     "condition",
	ident:         "ident",
	operator:      "op",
	LeftOperand:   "lhs",
	RightOperand:  "rhs",
}

func init() {
	for k, v := range parse.NodeNames {
		names[k] = v
	}

	parse.NodeNames = names
}

func Parse(source string) (*parse.Tree, error) {
	tree := parse.NewTree("rules", source, parseRoot)
	err := tree.Parse(Lex(source))
	return tree, err
}

func parseRoot(tree *parse.Tree) {
	tok := tree.Next()
	if tok.Type() == ActionIdentifier && tree.Peek().Type() == leftBrace {
		parseConditionList(tree, tok)
	}
	tree.Unexpected(tok, "ident")
}

//parseConditionList parses the rule for the action
func parseConditionList(tree *parse.Tree, nameToken lex.Token) {
	list := tree.AddNonTerminal(ConditionList, nameToken)
	list.AddTerminal(ident, nameToken)
	tree.Curr = list
	tree.Next() //This is a left brace, and has already been checked by parseRoot
	for {
		tok := tree.Next()
		switch tok.Type {
		case rightBrace:
			list.CommitSubTree()
			tree.Curr = tree.Root
			break
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
		default:
			tree.ErrorAtTokenf(tok, "unexpected %v, expected a rule", tok)
		}
	}
	parseRoot(tree)
}

func parseCondition(tree *parse.Tree, firstToken lex.Token) {
	condition := tree.AddNonTerminal(Condition, firstToken)
	tree.Curr = condition

	lhs := tree.AddNonTerminal(LHS, firstToken)
	tree.Curr = lhs
	parseExpression(tree, firstToken)
	tree.Curr = condition

	booleanOp := false

	switch tok := tree.Next(); tok.Type() {
	case EQ, NE, GT, LT, GE, LE:
		parseArithmeticOperator(tree, tok)
	case AND, OR:
		booleanOp = true
		parseBooleanOperator(tree, tok)
	default:
		tree.ErrorAtTokenf(tok, "%v is not an operator", tok)
	}

	switch rhtok := tree.Next(); rhtok.Type() {
	case name, leftParen:
		rhs := tree.AddNonTerminal(RHS, rhtok)
		tree.Curr = rhs
		if booleanOp {
			parseExpression(tree, rhtok)
		} else {
			parseIdent(tree, rhtok)
		}
	default:
		tree.ErrorAtTokenf(tok, "%v is not an operator", tok)
	}
	tree.Curr = condition
	//lhs op rhs has been parsed
	switch tok := tree.Next(); tok.Type() {
	case name:
		if tree.NestLevel > 0 {
			tree.ErrorAtTokenf(tok, "unclosed (")
		}
		tree.Back()
		//Return to conditionlist
		return
	case rightParen:
		if tree.NestLevel <= 0 {
			tree.ErrorAtTokenf(tok, "unexpected )")
		}
		tree.NestLevel--
	case rightBrace:
		if tree.NestLevel > 0 {
			tree.ErrorAtTokenf(tok, "unclosed (")
		}
	case AND, OR:
		//TODO:special case, slurp left side into the lhs of this operation
		//1. detach the already parsed operation from the tree
		//2. add a new Condition
		//2. add a new LHS
		//3. add the operation to the LHS
		//4. continue parsing the operator
	}
}

func parseExpression(tree *parse.Tree, firstToken lex.Token) {
	switch firstToken.Type() {
	case leftParen:
		a := tree.Next()
		b := tree.Next()
		if a.Type() == name && b.Type() == rightParen {
			tree.AddTerminal(ident, a)
			return
		}
		tree.NestLevel++
		parseCondition(tree.Back())
	case name:
		tree.AddTerminal(ident, firstToken)
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
			tree.AddTerminal(ident, a)
			return
		}
		tree.ErrorAtTokenf(firstToken, "unexpected %v, expected an identifier", firstToken)
	case name:
		tree.AddTerminal(ident, firstToken)
	default:
		tree.ErrorAtTokenf(firstToken, "unexpected %v, expected an identifier", firstToken)
	}
}

func parseArithmeticOperator(tree *parse.Tree, opToken lex.Token) {
	//TODO:check that the previous lhs is a an integer value
	//or leave this to the next pass
	tree.AddTerminal(operator, opToken)
}

func parseBooleanOperator(tree *parse.Tree, opToken lex.Token) {
	//TODO:check that the previous lhs is a boolean expression (a condition)
	//or leave this to the next pass
	tree.AddTerminal(operator, opToken)
	//with bool ops, the previous condition may be slurped into the lhs of this one
}
