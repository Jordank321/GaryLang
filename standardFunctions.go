package main

var standardFunctions map[string]functionDefinitionTree
var externDependencies map[string][]string
var setup bool = false

func GetStandardFunction(function string) functionDefinitionTree {
	if !setup {
		setupStandardFunctions()
	}
	return standardFunctions[function]
}
func GetStandardFunctionExterns(function string) []string {
	if !setup {
		setupStandardFunctions()
	}
	return externDependencies[function]
}

func setupStandardFunctions() {
	standardFunctions = map[string]functionDefinitionTree{
		"printthething": functionDefinitionTree{
			parameters:        []string{"printString"},
			assembledBodyFile: getAdr("printf"),
		},
	}
	externDependencies = map[string][]string{
		"printf": []string{
			"printf",
		},
	}
}
