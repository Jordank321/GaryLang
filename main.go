package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/Jordank321/GaryLang/asmFiles"
)

//go:generate go run scripts/includeasm.go

var nextParamNumber = 0

func main() {
	filePath := os.Args[1]
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	tokens := tokenize(string(fileBytes))
	tree := treeFromTokens(tokens)
	body := getAssemblyBodyFromTree(tree)
	asmFiles := usedBuiltinFunctions(tree, &[]string{})
	externs := cExternsFromAssemblyFiles(*asmFiles)
	consts := getAssemblyConstantsFromTree(tree)

	asmContents := getAssembly(body, externs, *asmFiles, consts)

	dir, name := path.Split(filePath)

	asmPath := dir + strings.Replace(name, ".gry", ".asm", -1)
	ioutil.WriteFile(asmPath, []byte(asmContents), os.ModeExclusive)

	objPath := dir + strings.Replace(name, ".gry", ".obj", -1)
	exec.Command("nasm", asmPath, "-fwin64", "-o"+objPath).Run()

	exePath := dir + strings.Replace(name, ".gry", ".exe", -1)
	exec.Command("gcc", objPath, "-m64", "-o"+exePath).Run()
}

func getAssembly(body string, externImports []string, builtInAsmFunctions []string, consts map[string][]byte) string {
	content := asmFiles.Start64bit
	content += asmFiles.Datasection
	content += constantsAsAsmString(consts)
	content += asmFiles.Alignconstbytes
	content += asmFiles.Codesection
	content += asmFiles.Invoke
	content += asmFiles.C
	for _, extern := range externImports {
		content += "extern " + extern + "\n"
	}
	content += "\nmain:\n"
	content += asmFiles.Allocatestack
	content += "\n" + body + "\n"
	content += asmFiles.Releasestack
	content += asmFiles.Exit
	return content
}

func constantsAsAsmString(consts map[string][]byte) string {
	result := ""
	for name, bytes := range consts {
		if bytes[len(bytes)-1] == 0 {
			result += name + ": db \"" + string(bytes[:len(bytes)-1]) + "\",0\n"
		} else {
			result += name + ": db \"" + string(bytes) + "\"\n"
		}
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
				constName := call.paramConstNames[parm]
				currentConstants[constName] = call.parameters[parm].evalValue
			}
		}
	}
	return currentConstants
}

func getAssemblyBodyFromTree(tree functionCallTree) string {
	currentBody := ""
	initBody := tree.definition.body
	for _, call := range initBody {
		if call.definition.assembledBodyFile != nil {
			assembly := *call.definition.assembledBodyFile
			for paramName, constName := range call.paramConstNames {
				assembly = strings.Replace(assembly, "$"+paramName, "$"+constName, -1)
			}
			currentBody += assembly
		}
	}
	return currentBody
}

// func readAsmFile(file string) string {
// 	contents, err := ioutil.ReadFile("./windowsAssembly/" + file + ".asm")
// 	if err != nil {
// 		panic(err)
// 	}
// 	return strings.Replace(string(contents), "\r\n", "\n", -1)
// }

func usedBuiltinFunctions(tree functionCallTree, used *[]string) *[]string {
	for _, param := range tree.parameters {
		usedBuiltinFunctions(param, used)
	}
	if tree.definition == nil {
		return used
	}
	asmFile := tree.definition.assembledBodyName
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
				definition:      &def,
				parameters:      make(map[string]functionCallTree),
				paramConstNames: make(map[string]string),
			}
		}
		if tokenCur.Type == paramOpen && inBody && !inParams {
			inParams = true
			continue
		}
		if inBody && inParams && tokenCur.Type == stringConst {
			paramName := callCur.definition.parameters[callCurParamNumber]
			callCurParamNumber++
			callCur.parameters[paramName] = functionCallTree{evalValue: append([]byte(*tokenCur.Value), 0)}
			callCur.paramConstNames[paramName] = "p" + strconv.Itoa(nextParamNumber)
			nextParamNumber++
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
	assembledBodyName *string
	assembledBodyFile *string
}

type functionCallTree struct {
	definition      *functionDefinitionTree
	evalValue       []byte
	parameters      map[string]functionCallTree
	paramConstNames map[string]string
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
