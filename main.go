package main

import (
	"C_Compiler/emitter"
	"C_Compiler/lexer"
	"C_Compiler/parser"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Error: Compiler needs source file as argument.")
		os.Exit(1)
	}

	source, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("Error: Could not read file: " + os.Args[1])
		os.Exit(1)
	}

	l := lexer.NewLexer(string(source))
	e := emitter.NewEmitter("out.asm")
	p := parser.NewParser(l, e)

	p.Program()
	e.WriteFile()
	fmt.Println("Compiling completed.")
}
