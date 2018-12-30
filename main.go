package main

import (
	"flag"
	"io/ioutil"
	"strings"
)

func main() {
	filePath := flag.Arg(0)
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	tokenize(string(fileBytes))
	tree := treeFromTokens(tokens)
	//assembly := writeToAssembly(tree)
	//binary := compileAssembly(assembly)
	//writeExecutable(binary)
}

func treeFromTokens(tokens *[]token) map[string]functionTree{
	return map[string]functionTree{

	}
}

func tokenize(input string) *[]token {
	tokens := []token{}
	lines := strings.Split(strings.Replace(input, "\r\n", "\n", -1), "\n")
	for _, line := range lines {
		words := strings.Split(line, " ")
		for _, word := range words {
			tok := parseWordToToken(strings.TrimSpace(word))
			tokens = append(tokens, tok)
		}
	}
	return &tokens
}

func parseWordToToken(input string) token {
	tok := token{}

	switch input {
	case "halfleft":
		tok.Type = procedureDefine
	case "alien":
		tok.Type = moduleImport
	case "£":
		tok.Type = paramOpen
	case "$":
		tok.Type = paramClose
	case "#":
		tok.Type = endLine
	case "/":
		tok.Type = bodyStart
	case "\\":
		tok.Type = bodyEnd
	default:
		if input[0] == '¬' || input[len(input)-1] == '¬' {
			tok.Type = stringConst
			tok.Value = getAdr(input[2 : len(input)-2])
		} else {
			tok.Type = name
			tok.Value = getAdr(input)
		}
	}

	return tok
}

type functionDefinitionTree struct {
	parameters map[string]string
	returnTypes map[string]string
	body []action
}

type functionCallTree struct{
	definition *functionDefinitionTree
	parameters map[string]func
}

type token struct {
	Type  tokeType
	Value *string
}

type tokeType int

const (
	moduleImport tokeType = iota
	procedureDefine
	name
	paramOpen
	paramClose
	endLine
	bodyStart
	bodyEnd
	stringConst
	eof
)

func getAdr(input string) *string {
	return &input
}
