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
	tokens := tokenize(string(fileBytes))
	treeFromTokens(tokens)
	//assembly := writeToAssembly(tree)
	//binary := compileAssembly(assembly)
	//writeExecutable(binary)
}

func treeFromTokens(tokens *[]token) functionCallTree {
	groups := map[string]*[]token{}
	currentFuncGroup := []token{}
	for _, tokenCur := range *tokens {
		if tokenCur.Type == procedureDefine {
			if len(currentFuncGroup) > 0 {
				groups[*currentFuncGroup[1].Value] = &currentFuncGroup
			}
			currentFuncGroup := []token{}
			currentFuncGroup = append(currentFuncGroup, tokenCur)
			continue
		} else if len(currentFuncGroup) > 0 {
			currentFuncGroup = append(currentFuncGroup, tokenCur)
		}

	}

	definitions := map[string]functionDefinitionTree{}
	for procName, procTokens := range groups {
		definitions[procName] = funcTree(procTokens)
	}

	initFunc := definitions["thisisthepie"]
	return functionCallTree{
		definition: &initFunc,
	}
}

func funcTree(tokens *[]token) functionDefinitionTree {
	tree := functionDefinitionTree{}

	inBody := false

	var paramCur *functionCallTree

	params := map[string]functionCallTree{}
	for _, tokenCur := range *tokens {
		if paramCur != nil {
			if tokenCur.Type == name {
				pa
			}
		}
		if tokenCur.Type == paramOpen && inBody == false && paramCur == nil {
			inParams = true
		}
	}

	return tree
}

func tokenize(input string) *[]token {
	tokens := []token{}
	lines := strings.Split(strings.Replace(input, "\r\n", "\n", -1), "\n")
	for _, line := range lines {
		words := strings.Split(line, " ")
		var stringTok *token
		for _, word := range words {
			if len(word) > 1 {
				if word[1] == '¬' && word[len(word)-1] != '¬' {
					stringTok = &token{
						Type:  stringConst,
						Value: getAdr(word[2:] + " "),
					}
					continue
				} else if word[1] != '¬' && word[len(word)-1] == '¬' {
					stringTok.Value = getAdr(*stringTok.Value + word[:len(word)-2])
					tokens = append(tokens, *stringTok)
					continue
				} else if stringTok != nil {
					stringTok.Value = getAdr(*stringTok.Value + word + " ")
					continue
				}
			}
			if len(word) == 0 && stringTok != nil {
				stringTok.Value = getAdr(*stringTok.Value + " ")
				continue
			}

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
		if input[1] == '¬' && input[len(input)-1] == '¬' {
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
	parameters  map[string]functionCallTree
	returnTypes map[string]string
	body        []functionCallTree
}

type functionCallTree struct {
	definition *functionDefinitionTree
	evalValue  *string
	parameters map[string]functionCallTree
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
