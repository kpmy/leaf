package rt

import (
	"leaf/ir"
	"leaf/ir/operation"
	"leaf/ir/types"
)

const StdDef = ` (* Builtin procedures definition *)
DEFINITION STD

	PROCEDURE NEW
	VAR p+ PTR
	END NEW

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

	PROCEDURE TYPEOF
	VAR
		res+ ATOM
		in- ANY
	INFIX res in
	END TYPEOF

	PROCEDURE INCL
		VAR set+ SET
		VAR x- ANY
	PRE (x # UNDEF)
	END INCL

	PROCEDURE EXCL
		VAR set+ SET
		VAR x- ANY
	PRE (x # UNDEF)
	END EXCL

	PROCEDURE VALUES
		VAR x- ANY
		VAR out+ LIST
	INFIX out x
	END VALUES

	PROCEDURE KEYS
		VAR x- MAP
		VAR out+ LIST
	INFIX out x
	END KEYS

	PROCEDURE PROCESS
		VAR to+, from+ MAP
	END PROCESS

	PROCEDURE ASSERT
		VAR cond- BOOLEAN; msg- ANY; code- INTEGER
	END ASSERT

	PROCEDURE HALT
		VAR code- INTEGER; msg- ANY
	END HALT

	PROCEDURE RUN
		VAR proc- PROCEDURE
		PRE proc # UNDEF
	END RUN
END STD.
`

type Qualident struct {
	Mod  string
	Proc string
}

type Message map[interface{}]interface{}

type Context interface {
	Handler() func(Message) Message
	Queue(interface{}, ...VarPar)
}

type Storage interface {
	List() []*ir.Variable
	Set(string, interface{})
	Get(string) interface{}
}

type Prop struct {
	Variadic bool
}

type VarPar struct {
	Name string
	Val  interface{}
	Sel  ir.Selector
}

type Calc func(types.Type, interface{}, operation.Operation, types.Type, interface{}, types.Type) interface{}
type Proc func(Context, Storage, Calc, ...VarPar)

var StdImp *ir.Import
var StdProc map[Qualident]Proc
var Special map[Qualident]Prop

const HANDLE = "HANDLE"
const MSG = "msg"

func init() {
	StdImp = &ir.Import{}
	StdImp.Init()
	StdProc = make(map[Qualident]Proc)
	Special = make(map[Qualident]Prop)
}
