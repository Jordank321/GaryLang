package main

import (
	"reflect"
	"testing"
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
								assembledBodyFile: getAdr("printf"),
								parameters: []string{
									"printString",
								},
							},
							parameters: map[string]functionCallTree{
								"printString": functionCallTree{
									evalValue: []byte("Hello  world!"),
								},
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
									assembledBodyFile: getAdr("printf"),
									parameters: []string{
										"printString",
									},
								},
								parameters: map[string]functionCallTree{
									"printString": functionCallTree{
										evalValue: []byte("Hello  world!"),
									},
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
		tree        functionCallTree
		currentBody string
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
									assembledBodyFile: getAdr("printf"),
									parameters: []string{
										"printString",
									},
								},
								parameters: map[string]functionCallTree{
									"printString": functionCallTree{
										evalValue: []byte("Hello  world!"),
									},
								},
							},
						},
					},
				},
				currentBody: "",
			},
			`; -----------------------------------------------------------------------------
; Call printf with seven parameters
; 4x of them are assigned to registers.
; 3x of them are assigned to stack spaces.
; -----------------------------------------------------------------------------
; Call printf with seven parameters
; -----------------------------------------------------------------------------
Invoke printf,$printString`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAssemblyBodyFromTree(tt.args.tree, tt.args.currentBody); got != tt.want {
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
										evalValue: []byte("Hello  world!"),
									},
								},
							},
						},
					},
				},
			},
			map[string][]byte{
				"printString": []byte("Hello  world!"),
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
