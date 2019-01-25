package main

import (
	"github.com/Jordank321/GaryLang/asmFiles"
)

var standardFunctions map[string]*FunctionDefinitionTree
var externDependencies map[string][]string
var setup bool

func GetStandardFunction(function string) *FunctionDefinitionTree {
	setupStandardFunctions()
	return standardFunctions[function]
}
func GetStandardFunctionExterns(function string) []string {
	setupStandardFunctions()
	return externDependencies[function]
}

func setupStandardFunctions() {
	if setup {
		return
	}
	standardFunctions = map[string]*FunctionDefinitionTree{
		"printthething": &FunctionDefinitionTree{
			Parameters:        []string{"printString"},
			AssembledBodyName: getAdr("printf"),
			AssembledBodyFile: getAdr(asmFiles.Printf),
		},
		"assign": &FunctionDefinitionTree{
			Parameters:        []string{"varName", "value", "valLength"},
			AssembledBodyName: getAdr("setbytes"),
			AssembledBodyFile: getAdr(asmFiles.Setbytes),
		},
	}
	externDependencies = map[string][]string{
		"printf": []string{
			"printf",
		},
		"setbytes": []string{
			"malloc",
			"free",
		},
	}
}
