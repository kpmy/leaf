@echo off
go build leaf.go
leaf -b="Log Core Objects HostObjects HostConsole Init TestEmpty TestFib TestFact TestBubble TestObjects TestEvents"
pause