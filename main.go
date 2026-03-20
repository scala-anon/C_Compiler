package main

import (
	"fmt"
	"C_Compiler/lexer"
)

func main() {
	source := "IF+-123 foo*THEN/"
	l := lexer.NewLexer(source)

	token := l.GetToken()

	for token.Kind != lexer.EOF {
		fmt.Println(token.Kind)
		token = l.GetToken()
	}
}
