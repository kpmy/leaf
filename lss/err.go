package lss

import (
	"encoding/xml"
	"fmt"
)

type Error struct {
	XMLName xml.Name
	From    string `xml:"from,attr"`
	Pos     int    `xml:"pos,attr"`
	Line    int    `xml:"line,attr"`
	Column  int    `xml:"column,attr"`
	Message string `xml:",chardata"`
}

func (e *Error) String() string {
	data, _ := xml.Marshal(e)
	return string(data)
}
func Err(sender string, pos, line, col int, msg ...interface{}) *Error {
	err := &Error{From: sender, Pos: pos, Line: line, Column: col, Message: fmt.Sprint(msg...)}
	err.XMLName.Local = "error"
	return err
}
