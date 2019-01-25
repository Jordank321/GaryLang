package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/Jordank321/GaryLang/asmFiles"
)

func Test_tokenize(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want *[]Token
	}{
		{
			"Initial Example",
			args{
				`halfleft thisisthepie £ $ /
	pie = 3 #
	printthething £ ¬Hello  world!¬ $ #
\`,
			},
			&[]Token{
				Token{
					Type: ProcedureDefine,
				},
				Token{
					Type:  Name,
					Value: getAdr("thisisthepie"),
				},
				Token{
					Type: ParamOpen,
				},
				Token{
					Type: ParamClose,
				},
				Token{
					Type: BodyStart,
				},
				Token{
					Type:  Name,
					Value: getAdr("pie"),
				},
				Token{
					Type: Assign,
				},
				Token{
					Type:  Number,
					Value: getAdr("3"),
				},
				Token{
					Type: EndLine,
				},
				Token{
					Type:  Name,
					Value: getAdr("printthething"),
				},
				Token{
					Type: ParamOpen,
				},
				Token{
					Type:  StringConst,
					Value: getAdr("Hello  world!"),
				},
				Token{
					Type: ParamClose,
				},
				Token{
					Type: EndLine,
				},
				Token{
					Type: BodyEnd,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tokenize(tt.args.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tokenize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_treeFromTokens(t *testing.T) {
	type args struct {
		tokens *[]Token
	}
	tests := []struct {
		name string
		args args
		want FunctionCallTree
	}{
		{
			"Initial Example",
			args{
				&[]Token{
					Token{
						Type: ProcedureDefine,
					},
					Token{
						Type:  Name,
						Value: getAdr("thisisthepie"),
					},
					Token{
						Type: ParamOpen,
					},
					Token{
						Type: ParamClose,
					},
					Token{
						Type: BodyStart,
					},
					Token{
						Type:  Name,
						Value: getAdr("pie"),
					},
					Token{
						Type: Assign,
					},
					Token{
						Type:  Number,
						Value: getAdr("3"),
					},
					Token{
						Type: EndLine,
					},
					Token{
						Type:  Name,
						Value: getAdr("printthething"),
					},
					Token{
						Type: ParamOpen,
					},
					Token{
						Type:  StringConst,
						Value: getAdr("Hello  world!"),
					},
					Token{
						Type: ParamClose,
					},
					Token{
						Type: EndLine,
					},
					Token{
						Type: BodyEnd,
					},
				},
			},
			FunctionCallTree{
				Definition: &FunctionDefinitionTree{
					Body: []FunctionCallTree{
						FunctionCallTree{
							Definition: &FunctionDefinitionTree{
								AssembledBodyName: getAdr("setbytes"),
								AssembledBodyFile: getAdr(asmFiles.Setbytes),
								Parameters: []string{
									"varName",
									"value",
									"valLength",
								},
							},
							Parameters: map[string]FunctionCallTree{
								"varName": FunctionCallTree{
									EvalValue: []byte{0},
								},
								"value": FunctionCallTree{
									EvalValue: []byte("3"),
								},
								"valLength": FunctionCallTree{
									EvalValue: []byte{1},
								},
							},
							ParamConstNames: map[string]string{
								"varName":   "p0",
								"value":     "p1",
								"valLength": "p2",
							},
						},
						FunctionCallTree{
							Definition: &FunctionDefinitionTree{
								AssembledBodyName: getAdr("printf"),
								AssembledBodyFile: getAdr(asmFiles.Printf),
								Parameters: []string{
									"printString",
								},
							},
							Parameters: map[string]FunctionCallTree{
								"printString": FunctionCallTree{
									EvalValue: append([]byte("Hello  world!"), 0),
								},
							},
							ParamConstNames: map[string]string{
								"printString": "p3",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := treeFromTokens(tt.args.tokens)
			if !reflect.DeepEqual(got, tt.want) {
				gotStr, err := json.MarshalIndent(got, "", "	")
				if err != nil {
					panic(err)
				}
				wantStr, err := json.MarshalIndent(tt.want, "", "	")
				if err != nil {
					panic(err)
				}
				t.Errorf("treeFromTokens() = %s, want %s", string(gotStr), string(wantStr))
			}
		})
	}
}

func Test_usedBuiltinFunctions(t *testing.T) {
	type args struct {
		tree FunctionCallTree
		used *[]string
	}
	tests := []struct {
		name string
		args args
		want *[]string
	}{
		{
			"Initial Examples",
			args{
				tree: FunctionCallTree{
					Definition: &FunctionDefinitionTree{
						Body: []FunctionCallTree{
							FunctionCallTree{
								Definition: &FunctionDefinitionTree{
									AssembledBodyName: getAdr("setbytes"),
									AssembledBodyFile: getAdr(asmFiles.Setbytes),
									Parameters: []string{
										"varName",
										"value",
									},
								},
								Parameters: map[string]FunctionCallTree{
									"varName": FunctionCallTree{
										EvalValue: []byte("pie"),
									},
									"value": FunctionCallTree{
										EvalValue: []byte("3"),
									},
								},
								ParamConstNames: map[string]string{
									"varName": "p0",
									"value":   "p1",
								},
							},
							FunctionCallTree{
								Definition: &FunctionDefinitionTree{
									AssembledBodyName: getAdr("printf"),
									AssembledBodyFile: getAdr(asmFiles.Printf),
									Parameters: []string{
										"printString",
									},
								},
								Parameters: map[string]FunctionCallTree{
									"printString": FunctionCallTree{
										EvalValue: append([]byte("Hello  world!"), 0),
									},
								},
								ParamConstNames: map[string]string{
									"printString": "p2",
								},
							},
						},
					},
				},
				used: &[]string{},
			},
			&[]string{
				"setbytes",
				"printf",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := usedBuiltinFunctions(tt.args.tree, tt.args.used); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("usedBuiltinFunctions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getAssemblyBodyFromTree(t *testing.T) {
	type args struct {
		tree FunctionCallTree
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"Initial Example",
			args{
				tree: FunctionCallTree{
					Definition: &FunctionDefinitionTree{
						Body: []FunctionCallTree{
							FunctionCallTree{
								Definition: &FunctionDefinitionTree{
									AssembledBodyName: getAdr("setbytes"),
									AssembledBodyFile: getAdr(asmFiles.Setbytes),
									Parameters: []string{
										"varName",
										"value",
									},
								},
								Parameters: map[string]FunctionCallTree{
									"varName": FunctionCallTree{
										EvalValue: []byte("pie"),
									},
									"value": FunctionCallTree{
										EvalValue: []byte("3"),
									},
								},
								ParamConstNames: map[string]string{
									"varName": "p0",
									"value":   "p1",
								},
							},
							FunctionCallTree{
								Definition: &FunctionDefinitionTree{
									AssembledBodyName: getAdr("printf"),
									AssembledBodyFile: getAdr(asmFiles.Printf),
									Parameters: []string{
										"printString",
									},
								},
								Parameters: map[string]FunctionCallTree{
									"printString": FunctionCallTree{
										EvalValue: append([]byte("Hello  world!"), 0),
									},
								},
								ParamConstNames: map[string]string{
									"printString": "p2",
								},
							},
						},
					},
				},
			},
			`; -----------------------------------------------------------------------------
; Call printf with seven parameters
; 4x of them are assigned to registers.
; 3x of them are assigned to stack spaces.
; -----------------------------------------------------------------------------
; Call printf with seven parameters
; -----------------------------------------------------------------------------
Invoke printf,$p0
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAssemblyBodyFromTree(tt.args.tree); got != tt.want {
				fmt.Println(got)
				t.Errorf("getAssemblyBodyFromTree() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getAssemblyConstantsFromTree(t *testing.T) {
	type args struct {
		tree FunctionCallTree
	}
	tests := []struct {
		name string
		args args
		want map[string][]byte
	}{
		{
			"Initial Example",
			args{
				FunctionCallTree{
					Definition: &FunctionDefinitionTree{
						Body: []FunctionCallTree{
							FunctionCallTree{
								Definition: &FunctionDefinitionTree{
									AssembledBodyName: getAdr("setbytes"),
									AssembledBodyFile: getAdr(asmFiles.Setbytes),
									Parameters: []string{
										"varName",
										"value",
									},
								},
								Parameters: map[string]FunctionCallTree{
									"varName": FunctionCallTree{
										EvalValue: []byte("pie"),
									},
									"value": FunctionCallTree{
										EvalValue: []byte("3"),
									},
								},
								ParamConstNames: map[string]string{
									"varName": "p0",
									"value":   "p1",
								},
							},
							FunctionCallTree{
								Definition: &FunctionDefinitionTree{
									AssembledBodyName: getAdr("printf"),
									AssembledBodyFile: getAdr(asmFiles.Printf),
									Parameters: []string{
										"printString",
									},
								},
								Parameters: map[string]FunctionCallTree{
									"printString": FunctionCallTree{
										EvalValue: append([]byte("Hello  world!"), 0),
									},
								},
								ParamConstNames: map[string]string{
									"printString": "p2",
								},
							},
						},
					},
				},
			},
			map[string][]byte{
				"p0": []byte("pie"),
				"p1": []byte("3"),
				"p2": append([]byte("Hello  world!"), 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAssemblyConstantsFromTree(tt.args.tree); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAssemblyConstantsFromTree() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cExternsFromAssemblyFiles(t *testing.T) {
	type args struct {
		asmFiles []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"Initial Example",
			args{
				[]string{
					"setbytes",
					"printf",
				},
			},
			[]string{
				"printf",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cExternsFromAssemblyFiles(tt.args.asmFiles); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cExternsFromAssemblyFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getAssemblyBody(t *testing.T) {
	type args struct {
		body                string
		externImports       []string
		builtInAsmFunctions []string
		contants            map[string][]byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"Simple Example",
			args{
				`; -----------------------------------------------------------------------------
; Call printf with seven parameters
; 4x of them are assigned to registers.
; 3x of them are assigned to stack spaces.
; -----------------------------------------------------------------------------
; Call printf with seven parameters
; -----------------------------------------------------------------------------
Invoke printf,$printString
`,
				[]string{
					"printf",
				},
				[]string{
					"thisisbuiltin",
				},
				map[string][]byte{
					"printString": append([]byte("Something that is forever constant!"), 0),
				},
			},
			`; ---------------------------------------------------------------------------
; Tell compiler to generate 64 bit code
; ---------------------------------------------------------------------------
bits 64
; ---------------------------------------------------------------------------
; Data segment:
; ---------------------------------------------------------------------------
section .data use64
printString: db "Something that is forever constant!",0
align 16 ; align data constants to the 16 byte boundary
; ---------------------------------------------------------------------------
; Code segment:
; ---------------------------------------------------------------------------
section .text use64
; ---------------------------------------------------------------------------
; Define macro: Invoke
; ---------------------------------------------------------------------------
%macro Invoke 1-*
        %if %0 > 1
                %rotate 1
                mov rcx,qword %1
                %rotate 1
                %if %0 > 2
                        mov rdx,qword %1
                        %rotate 1
                        %if  %0 > 3
                                mov r8,qword %1
                                %rotate 1
                                %if  %0 > 4
                                        mov r9,qword %1
                                        %rotate 1
                                        %if  %0 > 5
                                                %assign max %0-5
                                                %assign i 32
                                                %rep max
                                                        mov rax,qword %1
                                                        mov qword [rsp+i],rax
                                                        %assign i i+8
                                                        %rotate 1
                                                %endrep
                                        %endif
                                %endif
                        %endif
                %endif
        %endif
        ; ------------------------
        ; call %1 ; would be the same as this:
        ; -----------------------------------------
        sub rsp,qword 8
        mov qword [rsp],%%returnAddress
        jmp %1
        %%returnAddress:
        ; -----------------------------------------
%endmacro
; ---------------------------------------------------------------------------
; C management
; ---------------------------------------------------------------------------
global main
extern printf

main:
; -----------------------------------------------------------------------------
; Allocate stack memory
; -----------------------------------------------------------------------------
sub rsp,8*7

; -----------------------------------------------------------------------------
; Call printf with seven parameters
; 4x of them are assigned to registers.
; 3x of them are assigned to stack spaces.
; -----------------------------------------------------------------------------
; Call printf with seven parameters
; -----------------------------------------------------------------------------
Invoke printf,$printString

; -----------------------------------------------------------------------------
; Release stack memory
; -----------------------------------------------------------------------------
add rsp,8*7
; -----------------------------------------------------------------------------
; Quit
; -----------------------------------------------------------------------------
mov rax,qword 0
ret

; ----
; END ----
; ----
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAssembly(tt.args.body, tt.args.externImports, tt.args.builtInAsmFunctions, tt.args.contants); got != tt.want {
				gotLines := strings.Split(got, "\n")
				wantLines := strings.Split(tt.want, "\n")
				if len(gotLines) != len(wantLines) {
					t.Error("Expected the same number of lines")
					return
				}
				for i := 0; i < len(gotLines); i++ {
					gotLine := gotLines[i]
					wantLine := wantLines[i]
					if gotLine != wantLine {
						t.Errorf("\n%v\nwant\n%v", gotLine, wantLine)
						return
					}
				}
				t.Errorf("getAssemblyBody() = %v, want %v", got, tt.want)
			}
		})
	}
}
