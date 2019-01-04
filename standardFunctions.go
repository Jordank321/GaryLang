package main

var standardFunctions map[string]functionDefinitionTree

func setupStandardFunctions() {
	standardFunctions = map[string]functionDefinitionTree{
		"printthething": functionDefinitionTree{
			parameters: []string{"printString"},
			assembledBody
		},
	}
}
