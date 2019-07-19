package language

import (
	"kugg/compilers/parse"
)

const (
	ConditionList parse.NodeType = iota
	Condition
	LeftOperand
	RightOperand
	Operator
)

//Tree will be like
//Root
//	ConditionList (attack)
//		Condition (1)
//			Operator (LT)
//				LHS (targetlevel)
//				RHS (sourcelevel)
//		Condition (2)
//			Operator (GT)
//				LHS (moverange)
//				RHS (dist)
//	ConditionList (move)
//		Condition (1)
//			Operator (OR)
//				Operator (GT)
//					LHS bla
//					RHS ble
//				Operator (GT)
//					LHS ble
//					RHS blu
//	ConditionList (build)
//		Condition
//			Operator
//				LHS
//				RHS
//		Condition
//

var names = map[parse.NodeType]string{
	ConditionList: "condition-list",
	Condition:     "condition",
	LeftOperand:   "lhs",
	RightOperand:  "rhs",
	Operator:      "op",
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

func parseRoot(tree *parse.Tree) parse.StateFn {
	tree.Next()
	return nil
}

func parseCondition() {

}
