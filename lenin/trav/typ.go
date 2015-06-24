package trav

import (
	"fmt"
	"math/big"
)

type Atom string

type Int struct {
	big.Int
}

func NewInt(x int64) (ret *Int) {
	ret = &Int{}
	ret.Int = *big.NewInt(x)
	return
}

func ThisInt(x *big.Int) (ret *Int) {
	ret = &Int{}
	ret.Int = *x
	return
}

func (i *Int) String() string {
	x, _ := i.Int.MarshalText()
	return string(x)
}

type Rat struct {
	big.Rat
}

func NewRat(x float64) (ret *Rat) {
	ret = &Rat{}
	ret.Rat = *big.NewRat(0, 1)
	return
}

func ThisRat(x *big.Rat) (ret *Rat) {
	ret = &Rat{}
	ret.Rat = *x
	return
}

type Cmp struct {
	re, im *big.Rat
}

func (c *Cmp) String() (ret string) {
	null := big.NewRat(0, 1)
	if c.re.Cmp(null) != 0 {
		ret = fmt.Sprint(c.re)
	}
	if eq := c.im.Cmp(null); eq > 0 {
		ret = fmt.Sprint(ret, "+i", c.im.Abs(c.im))
	} else if eq < 0 {
		ret = fmt.Sprint(ret, "-i", c.im.Abs(c.im))
	} else if ret == "" {
		ret = "0"
	}
	return
}

func NewCmp(re, im float64) (ret *Cmp) {
	ret = &Cmp{}
	ret.re = big.NewRat(0, 1).SetFloat64(re)
	ret.im = big.NewRat(0, 1).SetFloat64(im)
	return
}

func ThisCmp(c *Cmp) (ret *Cmp) {
	ret = &Cmp{}
	*ret = *c
	return
}
