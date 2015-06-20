package yt

type ExprType string

const (
	Atom          ExprType = "atom"
	Constant      ExprType = "constant"
	NamedConstant ExprType = "named-constant"
	Variable      ExprType = "variable"
	Monadic       ExprType = "monadic"
	Dyadic        ExprType = "dyadic"
)

type StmtType string

const (
	Assign StmtType = "assign"
)
