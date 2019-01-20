package main

import (
	"github.com/Jordank321/GaryLang/asmFiles"
)

var standardFunctions map[string]functionDefinitionTree
var externDependencies map[string][]string
var setup bool

func GetStandardFunction(function string) functionDefinitionTree {
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
	standardFunctions = map[string]functionDefinitionTree{
		"printthething": functionDefinitionTree{
			parameters:        []string{"printString"},
			assembledBodyName: getAdr("printf"),
			assembledBodyFile: getAdr(asmFiles.Printf),
		},
	}
	externDependencies = map[string][]string{
		"printf": []string{
			"printf",
		},
	}
}
