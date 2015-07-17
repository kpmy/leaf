@echo off
go build leaf.go
go test
leaf -b="Log Core Objects HostObjects HostConsole Init TestEmpty TestFib TestFact TestErast TestBubble TestObjects TestEvents TestHandler"
pause
