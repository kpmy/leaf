@echo off
go build leaf.go
leaf -b=Log
leaf -b=Core
leaf -b=HostConsole
leaf -b=Init
leaf -b=TestFib
pause