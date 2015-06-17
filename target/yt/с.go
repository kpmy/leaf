package yt

type ExprType string

const (
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
