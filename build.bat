@echo off
go build leaf.go
leaf -b="Log Core HostConsole Init TestEmpty TestFib TestFact TestBubble"
pause