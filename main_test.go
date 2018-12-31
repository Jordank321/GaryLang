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
