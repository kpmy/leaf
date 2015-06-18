package tri

/* троичная логика */
var TRUE Trit = Trit{N: false, T: true}
var FALSE Trit = Trit{N: false, T: false}
var NIL Trit = Trit{N: true, T: false}

type Trit struct {
	N bool
	T bool
}

func (t Trit) String() string {
	if t.N {
		return "%nil"
	} else if t.T {
		return "%true"
	} else {
		return "%false"
	}
}

func False(t Trit) bool {
	return t == FALSE
}

func True(t Trit) bool {
	return t == TRUE
}

func Nil(t Trit) bool {
	return t == NIL
}

func Not(t Trit) Trit {
	if t == TRUE {
		return FALSE
	} else if t == FALSE {
		return TRUE
	} else if t == NIL {
		return NIL
	}
	panic(0)
}

func Impl(p, q Trit) Trit {
	if False(p) && False(q) {
		return TRUE
	} else if False(p) && True(q) {
		return TRUE
	} else if True(p) && False(q) {
		return FALSE
	} else if True(p) && True(q) {
		return TRUE
	} else if True(p) && Nil(q) {
		return NIL
	} else if Nil(p) && False(q) {
		return NIL
	} else if False(p) && Nil(q) {
		return TRUE
	} else if Nil(p) && Nil(q) {
		return TRUE
	} else if Nil(p) && True(q) {
		return TRUE
	}
	panic(0)
}

func CNot(t Trit) Trit {
	if t == TRUE {
		return FALSE
	} else if t == FALSE {
		return NIL
	} else {
		return TRUE
	}
}

func Or(p, q Trit) Trit {
	return Impl(Impl(p, q), q)
}

func And(p, q Trit) Trit {
	return Not(Or(Not(p), Not(q)))
}

func Eq(p, q Trit) Trit {
	return And(Impl(p, q), Impl(q, p))
}

func This(_x interface{}) Trit {
	switch x := _x.(type) {
	case int:
		if x == 1 {
			return TRUE
		} else if x == 0 {
			return NIL
		} else if x == -1 {
			return FALSE
		}
	case bool:
		if x {
			return TRUE
		} else {
			return FALSE
		}
	case nil:
		return NIL
	}
	panic(0)
}

func Ord(t Trit) int {
	if t == FALSE {
		return -1
	} else if t == NIL {
		return 0
	} else if t == TRUE {
		return 1
	}
	panic(0)
}

func Sum3(p, q Trit) Trit {
	switch Ord(p) {
	case -1:
		return q
	case 0:
		if False(q) {
			return NIL
		} else if Nil(q) {
			return TRUE
		} else {
			return FALSE
		}
	case 1:
		if False(q) {
			return TRUE
		} else if Nil(q) {
			return FALSE
		} else {
			return NIL
		}
	default:
		panic(0)
	}
}

func Sum3r(p, q Trit) Trit {
	return CNot(CNot(Sum3(p, q)))
}

func CarryS(p, q Trit) Trit {
	switch Ord(p) {
	case -1:
		return FALSE
	case 0:
		if True(q) {
			return NIL
		} else {
			return FALSE
		}
	case 1:
		if False(q) {
			return FALSE
		} else {
			return NIL
		}
	default:
		panic(0)
	}
}

func CarrySr(p, q Trit) Trit {
	if False(p) && False(q) {
		return FALSE
	} else if True(p) && True(q) {
		return TRUE
	} else {
		return NIL
	}
}

func Mul3(p, q Trit) Trit {
	switch Ord(p) {
	case -1:
		return FALSE
	case 0:
		return q
	case 1:
		if False(q) {
			return FALSE
		} else if Nil(q) {
			return TRUE
		} else {
			return NIL
		}
	default:
		panic(0)
	}
}

func CarryM(p, q Trit) Trit {
	if True(p) && True(q) {
		return NIL
	} else {
		return FALSE
	}
}

func Mul3r(p, q Trit) Trit {
	if Nil(p) && Nil(q) {
		return NIL
	} else {
		if p == q {
			return TRUE
		} else {
			return FALSE
		}
	}
}

func Webb(p, q Trit) Trit {
	return CNot(Or(p, q))
}

func Mod(t Trit) Trit {
	return Or(t, Not(t))
}
