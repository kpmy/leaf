package yt

type ExprType string

const (
	Atom          ExprType = "atom"
	Constant      ExprType = "constant"
	NamedConstant ExprType = "named-constant"
	Variable      ExprType = "variable"
	Monadic       ExprType = "monadic"
	Dyadic        ExprType = "dyadic"
	SelExpr       ExprType = "selector"
)

type StmtType string

const (
	Assign StmtType = "assign"
)

type SelType string

const (
	SelVar SelType = "variable"
	SelIdx SelType = "index"
)
