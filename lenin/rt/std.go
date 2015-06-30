package rt

import (
	"leaf/ir"
	"leaf/ir/operation"
	"leaf/ir/types"
)

const StdDef = ` (* Builtin procedures definition *)
DEFINITION STD

	PROCEDURE INC
	VAR
		x+ INTEGER
		n- INTEGER
	PRE n >= 0
	END INC

	PROCEDURE DEC
	VAR
		x+ INTEGER
		n- INTEGER
	PRE n >= 0
	END DEC

	PROCEDURE CAP
	VAR
		cap+ CHAR
		x- CHAR
	INFIX cap x
	END CAP

	PROCEDURE LEN
	VAR
		in- ANY
		out+ INTEGER
	INFIX out in
	END LEN

	PROCEDURE ODD
	VAR
		in- INTEGER
		out+ BOOLEAN
	INFIX out in
	END ODD

	PROCEDURE RESIZE
	VAR
		list+ LIST
		n- INTEGER
	PRE n >= 0
	END RESIZE

END STD.
`

type Qualident struct {
	Mod  string
	Proc string
}

type Storage interface {
	List() []*ir.Variable
	Set(string, interface{})
	Get(string) interface{}
}

type Calc func(types.Type, interface{}, operation.Operation, types.Type, interface{}, types.Type) interface{}
type Proc func(Storage, Calc)

var StdImp *ir.Import
var StdProc map[Qualident]Proc

func init() {
	StdImp = &ir.Import{}
	StdImp.Init()
	StdProc = make(map[Qualident]Proc)
}
