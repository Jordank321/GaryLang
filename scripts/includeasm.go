package main

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// Reads all .txt files in the current folder
// and encodes them as strings literals in textfiles.go
func main() {
	fs, _ := ioutil.ReadDir("./windowsAssembly")
	out, _ := os.Create("./asmFiles/asmFiles.go")
	out.Write([]byte("package asmFiles \n\nconst (\n"))
	for _, f := range fs {
		if strings.HasSuffix(f.Name(), ".asm") {
			out.Write([]byte(strings.Title(strings.TrimSuffix(f.Name(), ".asm")) + " = `"))
			f, err := os.Open("./windowsAssembly/" + f.Name())
			if err != nil {
				panic(err)
			}
			io.Copy(out, f)
			out.Write([]byte("`\n"))
		}
	}
	out.Write([]byte(")\n"))
}
