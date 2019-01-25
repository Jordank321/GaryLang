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
	out, err := exec.Command("nasm", asmPath, "-fwin64", "-o"+objPath).Output()
	println(string(out))
	if err != nil {
		panic(err)
	}

	exePath := dir + strings.Replace(name, ".gry", ".exe", -1)
	out, err = exec.Command("gcc", objPath, "-m64", "-o"+exePath).Output()
	println(string(out))
	if err != nil {
		panic(err)
	}
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

func getAssemblyConstantsFromTree(tree FunctionCallTree) map[string][]byte {
	currentConstants := map[string][]byte{}
	initBody := tree.Definition.Body
	for _, call := range initBody {
		if call.Definition.AssembledBodyFile != nil {
			for _, parm := range call.Definition.Parameters {
				constName := call.ParamConstNames[parm]
				currentConstants[constName] = call.Parameters[parm].EvalValue
			}
		}
	}
	return currentConstants
}

func getAssemblyBodyFromTree(tree FunctionCallTree) string {
	currentBody := ""
	initBody := tree.Definition.Body
	for _, call := range initBody {
		if call.Definition.AssembledBodyFile != nil {
			assembly := *call.Definition.AssembledBodyFile
			for paramName, constName := range call.ParamConstNames {
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

func usedBuiltinFunctions(tree FunctionCallTree, used *[]string) *[]string {
	for _, param := range tree.Parameters {
		usedBuiltinFunctions(param, used)
	}
	if tree.Definition == nil {
		return used
	}
	asmFile := tree.Definition.AssembledBodyName
	if asmFile != nil {
		newUsed := appendIfMissing(*used, *asmFile)
		*used = newUsed
	}

	for _, call := range (*tree.Definition).Body {
		usedBuiltinFunctions(call, used)
	}
	return used
}

func treeFromTokens(tokens *[]Token) FunctionCallTree {
	groups := map[string]*[]Token{}
	var currentFuncGroup []Token
	for _, tokenCur := range *tokens {
		if tokenCur.Type == ProcedureDefine {
			if len(currentFuncGroup) > 0 {
				groups[*currentFuncGroup[1].Value] = &currentFuncGroup
			}
			currentFuncGroup = []Token{tokenCur}
			continue
		} else if len(currentFuncGroup) > 0 {
			currentFuncGroup = append(currentFuncGroup, tokenCur)
		}
	}
	groups[*currentFuncGroup[1].Value] = &currentFuncGroup

	definitions := map[string]FunctionDefinitionTree{}
	for procName, procTokens := range groups {
		definitions[procName] = funcTree(procTokens)
	}

	initFunc := definitions["thisisthepie"]
	return FunctionCallTree{
		Definition: &initFunc,
	}
}

func funcTree(tokens *[]Token) FunctionDefinitionTree {
	setupStandardFunctions()
	tree := FunctionDefinitionTree{}

	inBody := false
	inParams := false
	var leftHandSide *string
	var rightHandSide *string
	var callCur *FunctionCallTree
	callCurParamNumber := 0

	assignParam := func(param string) {
		callCur.ParamConstNames[param] = "p" + strconv.Itoa(nextParamNumber)
		nextParamNumber++
	}

	for _, tokenCur := range *tokens {
		if tokenCur.Type == ParamOpen && !inBody && !inParams {
			inParams = true
			continue
		}
		if tokenCur.Type == ParamClose && !inBody && inParams {
			inParams = false
			continue
		}
		if inParams && tokenCur.Type == Name {
			tree.Parameters = append(tree.Parameters, *tokenCur.Value)
			continue
		}

		if tokenCur.Type == BodyStart && !inBody {
			inBody = true
			continue
		}
		if tokenCur.Type == BodyEnd && inBody {
			inBody = true
			continue
		}

		if inBody && !inParams && tokenCur.Type == Name {
			def := standardFunctions[*tokenCur.Value]
			if def == nil {
				if leftHandSide == nil {
					leftHandSide = tokenCur.Value
				} else {
					rightHandSide = tokenCur.Value
				}
			} else {
				callCur = &FunctionCallTree{
					Definition:      def,
					Parameters:      map[string]FunctionCallTree{},
					ParamConstNames: map[string]string{},
				}
			}
		}
		if inBody && !inParams && tokenCur.Type == Assign {
			def := GetStandardFunction("assign")
			callCur = &FunctionCallTree{
				Definition:      def,
				Parameters:      map[string]FunctionCallTree{},
				ParamConstNames: map[string]string{},
			}
		}
		if inBody && !inParams && tokenCur.Type == Number && leftHandSide != nil {
			rightHandSide = tokenCur.Value
		}
		if inBody && !inParams && tokenCur.Type == StringConst && leftHandSide != nil {
			rightHandSide = tokenCur.Value
		}
		if tokenCur.Type == ParamOpen && inBody && !inParams {
			inParams = true
			continue
		}
		if inBody && inParams && tokenCur.Type == StringConst {
			paramName := callCur.Definition.Parameters[callCurParamNumber]
			callCurParamNumber++
			callCur.Parameters[paramName] = FunctionCallTree{EvalValue: append([]byte(*tokenCur.Value), 0)}
			assignParam(paramName)
		}
		if tokenCur.Type == ParamClose && inBody && inParams {
			inParams = false
			tree.Body = append(tree.Body, *callCur)
			callCur = nil
			callCurParamNumber = 0
			continue
		}

		if leftHandSide != nil && callCur != nil && rightHandSide != nil {
			lhsName := callCur.Definition.Parameters[0]
			rhsName := callCur.Definition.Parameters[1]
			callCur.Parameters[lhsName] = FunctionCallTree{EvalValue: []byte{0}}
			assignParam(lhsName)
			callCur.Parameters[rhsName] = FunctionCallTree{EvalValue: []byte(*rightHandSide)}
			assignParam(rhsName)
			callCur.Parameters["valLength"] = FunctionCallTree{EvalValue: []byte{byte(len(*rightHandSide))}}
			assignParam("valLength")
			tree.Body = append(tree.Body, *callCur)
			callCur = nil
			leftHandSide = nil
			rightHandSide = nil
			continue
		}
	}

	return tree
}

func tokenize(input string) *[]Token {
	tokens := []Token{}
	lines := strings.Split(strings.Replace(input, "\r\n", "\n", -1), "\n")
	for _, line := range lines {
		words := strings.Split(line, " ")
		var stringTok *Token
		for _, word := range words {
			if len(word) > 1 {
				if word[1] == '¬' && word[len(word)-1] != '¬' {
					stringTok = &Token{
						Type:  StringConst,
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

func parseWordToToken(input string) Token {
	tok := Token{}

	switch input {
	case "halfleft":
		tok.Type = ProcedureDefine
	case "alien":
		tok.Type = ModuleImport
	case "£":
		tok.Type = ParamOpen
	case "$":
		tok.Type = ParamClose
	case "#":
		tok.Type = EndLine
	case "/":
		tok.Type = BodyStart
	case "\\":
		tok.Type = BodyEnd
	case "=":
		tok.Type = Assign
	default:
		if len(input) >= 2 && input[1] == '¬' && input[len(input)-1] == '¬' {
			tok.Type = StringConst
			tok.Value = getAdr(input[2 : len(input)-2])
		} else if _, err := strconv.Atoi(input); err == nil {
			tok.Type = Number
			tok.Value = getAdr(input)
		} else {
			tok.Type = Name
			tok.Value = getAdr(input)
		}
	}

	return tok
}

type FunctionDefinitionTree struct {
	Parameters        []string
	Body              []FunctionCallTree
	AssembledBodyName *string
	AssembledBodyFile *string
}

type FunctionCallTree struct {
	Definition      *FunctionDefinitionTree
	EvalValue       []byte
	Parameters      map[string]FunctionCallTree
	ParamConstNames map[string]string
}

type Token struct {
	Type  TokeType
	Value *string
}

type TokeType int

const (
	ModuleImport TokeType = iota
	ProcedureDefine
	Name
	ParamOpen
	ParamClose
	EndLine
	BodyStart
	BodyEnd
	StringConst
	Assign
	Number
	EOF
)

func getAdr(input string) *string {
	return &input
}
