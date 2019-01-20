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

func getAssembly(body string, externImports []string, builtInAsmFunctions []string, consts map[string][]byte) string {
	content := readAsmFile("start64bit")
	content += readAsmFile("datasection")
	content += constantsAsAsmString(consts)
	content += readAsmFile("alignconstbytes")
	content += readAsmFile("codesection")
	content += readAsmFile("invoke")
	content += readAsmFile("c")
	for _, extern := range externImports {
		content += "extern " + extern + "\n"
	}
	content += "\nmain:\n"
	content += readAsmFile("allocatestack")
	content += "\n" + body + "\n"
	content += readAsmFile("releasestack")
	content += readAsmFile("exit")
	return content
}

func constantsAsAsmString(consts map[string][]byte) string {
	result := ""
	for name, bytes := range consts {
		result += name + ": db \"" + string(bytes) + "\"\n"
	}
	return result
}

func cExternsFromAssemblyFiles(asmFiles []string) []string {
	externs := []string{}
	for _, asmFile := range asmFiles {
		externs = append(externs, GetStandardFunctionExterns(asmFile)...)
	}
	return externs
}

func getAssemblyConstantsFromTree(tree functionCallTree) map[string][]byte {
	currentConstants := map[string][]byte{}
	initBody := tree.definition.body
	for _, call := range initBody {
		if call.definition.assembledBodyFile != nil {
			for _, parm := range call.definition.parameters {
				currentConstants[parm] = call.parameters[parm].evalValue
			}

		}
	}
	return currentConstants
}

func getAssemblyBodyFromTree(tree functionCallTree, currentBody string) string {
	initBody := tree.definition.body
	for _, call := range initBody {
		if call.definition.assembledBodyFile != nil {
			currentBody += readAsmFile(*call.definition.assembledBodyFile)
		}
	}
	return currentBody
}

func readAsmFile(file string) string {
	contents, err := ioutil.ReadFile("./windowsAssembly/" + file + ".asm")
	if err != nil {
		panic(err)
	}
	return strings.Replace(string(contents), "\r\n", "\n", -1)
}

func usedBuiltinFunctions(tree functionCallTree, used *[]string) *[]string {
	for _, param := range tree.parameters {
		usedBuiltinFunctions(param, used)
	}
	if tree.definition == nil {
		return used
	}
	asmFile := tree.definition.assembledBodyFile
	if asmFile != nil {
		newUsed := appendIfMissing(*used, *asmFile)
		*used = newUsed
	}

	for _, call := range (*tree.definition).body {
		usedBuiltinFunctions(call, used)
	}
	return used
}

func treeFromTokens(tokens *[]token) functionCallTree {
	groups := map[string]*[]token{}
	var currentFuncGroup []token
	for _, tokenCur := range *tokens {
		if tokenCur.Type == procedureDefine {
			if len(currentFuncGroup) > 0 {
				groups[*currentFuncGroup[1].Value] = &currentFuncGroup
			}
			currentFuncGroup = []token{tokenCur}
			continue
		} else if len(currentFuncGroup) > 0 {
			currentFuncGroup = append(currentFuncGroup, tokenCur)
		}
	}
	groups[*currentFuncGroup[1].Value] = &currentFuncGroup

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
	setupStandardFunctions()
	tree := functionDefinitionTree{}

	inBody := false
	inParams := false
	var callCur *functionCallTree
	callCurParamNumber := 0

	for _, tokenCur := range *tokens {
		if tokenCur.Type == paramOpen && !inBody && !inParams {
			inParams = true
			continue
		}
		if tokenCur.Type == paramClose && !inBody && inParams {
			inParams = false
			continue
		}
		if inParams && tokenCur.Type == name {
			tree.parameters = append(tree.parameters, *tokenCur.Value)
			continue
		}

		if tokenCur.Type == bodyStart && !inBody {
			inBody = true
			continue
		}
		if tokenCur.Type == bodyEnd && inBody {
			inBody = true
			continue
		}

		if inBody && !inParams && tokenCur.Type == name {
			def := standardFunctions[*tokenCur.Value]
			callCur = &functionCallTree{
				definition: &def,
				parameters: make(map[string]functionCallTree),
			}
		}
		if tokenCur.Type == paramOpen && inBody && !inParams {
			inParams = true
			continue
		}
		if inBody && inParams && tokenCur.Type == stringConst {
			paramName := callCur.definition.parameters[callCurParamNumber]
			callCurParamNumber++
			callCur.parameters[paramName] = functionCallTree{evalValue: []byte(*tokenCur.Value)}
		}
		if tokenCur.Type == paramClose && inBody && inParams {
			inParams = false
			tree.body = append(tree.body, *callCur)
			callCur = nil
			callCurParamNumber = 0
			continue
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
	parameters        []string
	body              []functionCallTree
	assembledBodyFile *string
}

type functionCallTree struct {
	definition *functionDefinitionTree
	evalValue  []byte
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
