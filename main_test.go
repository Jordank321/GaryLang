package main

import (
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
		want *[]token
	}{
		{
			"Initial Example",
			args{
				`halfleft thisisthepie £ $ /
	printthething £ ¬Hello  world!¬ $ #
\`,
			},
			&[]token{
				token{
					Type: procedureDefine,
				},
				token{
					Type:  name,
					Value: getAdr("thisisthepie"),
				},
				token{
					Type: paramOpen,
				},
				token{
					Type: paramClose,
				},
				token{
					Type: bodyStart,
				},
				token{
					Type:  name,
					Value: getAdr("printthething"),
				},
				token{
					Type: paramOpen,
				},
				token{
					Type:  stringConst,
					Value: getAdr("Hello  world!"),
				},
				token{
					Type: paramClose,
				},
				token{
					Type: endLine,
				},
				token{
					Type: bodyEnd,
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
		tokens *[]token
	}
	tests := []struct {
		name string
		args args
		want functionCallTree
	}{
		{
			"Initial Example",
			args{
				&[]token{
					token{
						Type: procedureDefine,
					},
					token{
						Type:  name,
						Value: getAdr("thisisthepie"),
					},
					token{
						Type: paramOpen,
					},
					token{
						Type: paramClose,
					},
					token{
						Type: bodyStart,
					},
					token{
						Type:  name,
						Value: getAdr("printthething"),
					},
					token{
						Type: paramOpen,
					},
					token{
						Type:  stringConst,
						Value: getAdr("Hello  world!"),
					},
					token{
						Type: paramClose,
					},
					token{
						Type: endLine,
					},
					token{
						Type: bodyEnd,
					},
				},
			},
			functionCallTree{
				definition: &functionDefinitionTree{
					body: []functionCallTree{
						functionCallTree{
							definition: &functionDefinitionTree{
								assembledBodyName: getAdr("printf"),
								assembledBodyFile: getAdr(asmFiles.Printf),
								parameters: []string{
									"printString",
								},
							},
							parameters: map[string]functionCallTree{
								"printString": functionCallTree{
									evalValue: append([]byte("Hello  world!"), 0),
								},
							},
							paramConstNames: map[string]string{
								"printString": "p0",
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
				t.Errorf("treeFromTokens() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_usedBuiltinFunctions(t *testing.T) {
	type args struct {
		tree functionCallTree
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
				tree: functionCallTree{
					definition: &functionDefinitionTree{
						body: []functionCallTree{
							functionCallTree{
								definition: &functionDefinitionTree{
									assembledBodyName: getAdr("printf"),
									assembledBodyFile: getAdr(asmFiles.Printf),
									parameters: []string{
										"printString",
									},
								},
								parameters: map[string]functionCallTree{
									"printString": functionCallTree{
										evalValue: []byte("Hello  world!"),
									},
								},
								paramConstNames: map[string]string{
									"printString": "p0",
								},
							},
						},
					},
				},
				used: &[]string{},
			},
			&[]string{
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
		tree functionCallTree
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"Initial Example",
			args{
				tree: functionCallTree{
					definition: &functionDefinitionTree{
						body: []functionCallTree{
							functionCallTree{
								definition: &functionDefinitionTree{
									assembledBodyName: getAdr("printf"),
									assembledBodyFile: getAdr(asmFiles.Printf),
									parameters: []string{
										"printString",
									},
								},
								parameters: map[string]functionCallTree{
									"printString": functionCallTree{
										evalValue: []byte("Hello  world!"),
									},
								},
								paramConstNames: map[string]string{
									"printString": "p0",
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
		tree functionCallTree
	}
	tests := []struct {
		name string
		args args
		want map[string][]byte
	}{
		{
			"Initial Example",
			args{
				functionCallTree{
					definition: &functionDefinitionTree{
						body: []functionCallTree{
							functionCallTree{
								definition: &functionDefinitionTree{
									assembledBodyFile: getAdr("printf"),
									parameters: []string{
										"printString",
									},
								},
								parameters: map[string]functionCallTree{
									"printString": functionCallTree{
										evalValue: append([]byte("Hello  world!"), 0),
									},
								},
								paramConstNames: map[string]string{
									"printString": "p0",
								},
							},
						},
					},
				},
			},
			map[string][]byte{
				"p0": append([]byte("Hello  world!"), 0),
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
