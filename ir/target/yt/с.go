package yt

type ExprType string

const (
	Atom          ExprType = "atom"
	Constant      ExprType = "constant"
	Set           ExprType = "set"
	List          ExprType = "list"
	Map           ExprType = "map"
	NamedConstant ExprType = "named-constant"
	Variable      ExprType = "variable"
	Monadic       ExprType = "monadic"
	Dyadic        ExprType = "dyadic"
	SelExpr       ExprType = "selector"
	Infix         ExprType = "infix"
	InvokeInfix   ExprType = "invoke"
	TypeTest      ExprType = "typetest"
)

type StmtType string

const (
	Assign StmtType = "assign"
	If     StmtType = "if"
	While  StmtType = "while"
	Repeat StmtType = "repeat"
	Choose StmtType = "choose"
	Call   StmtType = "call"
	Invoke StmtType = "invoke"
)

type SelType string

const (
	SelVar SelType = "variable"
	SelIdx SelType = "index"
	SelMod SelType = "module"
)
