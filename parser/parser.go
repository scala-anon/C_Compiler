package parser

import (
	"C_Compiler/lexer"
	"fmt"
	"os"
)

type Parser struct {
	Lexer     *lexer.Lexer
	CurToken  lexer.Token
	PeekToken lexer.Token
	Symbols   map[string]bool // Variables declared so far.
}

func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{
		Lexer:   l,
		Symbols: make(map[string]bool),
	}
	p.NextToken()
	p.NextToken()
	return p
}

// Return true if the current token matches.
func (p *Parser) CheckToken(kind lexer.TokenType) bool {
	return p.CurToken.Kind == kind
}

// Return true if the next token matches.
func (p *Parser) CheckPeek(kind lexer.TokenType) bool {
	return p.PeekToken.Kind == kind
}

// Try to match current token. If not, error. Advances the current token.
func (p *Parser) Match(kind lexer.TokenType) {
	if !p.CheckToken(kind) {
		p.Abort(fmt.Sprintf("Expected %d, got %d", kind, p.CurToken.Kind))
	}
	p.NextToken()
}

// Advances the current token.
func (p *Parser) NextToken() {
	p.CurToken = p.PeekToken
	p.PeekToken = p.Lexer.GetToken()
}

func (p *Parser) Abort(message string) {
	fmt.Println("Parse Error: " + message)
	os.Exit(1)
}

// program ::= {function | declaration}
func (p *Parser) Program() {
	fmt.Println("PROGRAM")

	for !p.CheckToken(lexer.EOF) {
		p.Function()
	}
}

// function ::= type ident "(" [params] ")" block
func (p *Parser) Function() {
	fmt.Println("FUNCTION")
	p.Type()
	p.Match(lexer.IDENT)
	p.Match(lexer.LPAREN)
	p.Params()
	p.Match(lexer.RPAREN)
	p.Block()
}

// params ::= type ident {"," type ident}
func (p *Parser) Params() {
	if !p.CheckToken(lexer.RPAREN) {
		p.Type()
		// Add param to symbols.
		p.Symbols[p.CurToken.Text] = true
		p.Match(lexer.IDENT)

		for p.CheckToken(lexer.COMMA) {
			p.NextToken()
			p.Type()
			p.Symbols[p.CurToken.Text] = true
			p.Match(lexer.IDENT)
		}
	}
}

// block ::= "{" {statement} "}"
func (p *Parser) Block() {
	p.Match(lexer.LBRACE)
	for !p.CheckToken(lexer.RBRACE) {
		p.Statement()
	}
	p.Match(lexer.RBRACE)
}

// type ::= "int" | "float" | "char" | "void"
func (p *Parser) Type() {
	switch p.CurToken.Kind {
	case lexer.INT, lexer.FLOAT, lexer.CHAR, lexer.VOID:
		p.NextToken()
	default:
		p.Abort("Expected type, got " + p.CurToken.Text)
	}
}

// statement router
func (p *Parser) Statement() {
	if p.CheckToken(lexer.IF) {
		p.IfStatement()
	} else if p.CheckToken(lexer.WHILE) {
		p.WhileStatement()
	} else if p.CheckToken(lexer.FOR) {
		p.ForStatement()
	} else if p.CheckToken(lexer.RETURN) {
		p.ReturnStatement()
	} else if p.CheckToken(lexer.INT) || p.CheckToken(lexer.FLOAT) || p.CheckToken(lexer.CHAR) || p.CheckToken(lexer.VOID) {
		p.DeclarationStatement()
	} else if p.CheckToken(lexer.IDENT) {
		p.IdentStatement()
	} else {
		p.Abort("Invalid statement: " + p.CurToken.Text)
	}
}

// "if" "(" comparison ")" block ["else" block]
func (p *Parser) IfStatement() {
	fmt.Println("STATEMENT-IF")
	p.Match(lexer.IF)
	p.Match(lexer.LPAREN)
	p.Comparison()
	p.Match(lexer.RPAREN)
	p.Block()
	if p.CheckToken(lexer.ELSE) {
		fmt.Println("STATEMENT-ELSE")
		p.NextToken()
		p.Block()
	}
}

// "while" "(" comparison ")" block
func (p *Parser) WhileStatement() {
	fmt.Println("STATEMENT-WHILE")
	p.Match(lexer.WHILE)
	p.Match(lexer.LPAREN)
	p.Comparison()
	p.Match(lexer.RPAREN)
	p.Block()
}

// "for" "(" declaration ";" comparison ";" ident "=" expression ")" block
func (p *Parser) ForStatement() {
	fmt.Println("STATEMENT-FOR")
	p.Match(lexer.FOR)
	p.Match(lexer.LPAREN)
	p.Declaration()
	p.Match(lexer.SEMICOLON)
	p.Comparison()
	p.Match(lexer.SEMICOLON)
	p.Match(lexer.IDENT)
	p.Match(lexer.EQ)
	p.Expression()
	p.Match(lexer.RPAREN)
	p.Block()
}

// "return" [expression] ";"
func (p *Parser) ReturnStatement() {
	fmt.Println("STATEMENT-RETURN")
	p.Match(lexer.RETURN)
	if !p.CheckToken(lexer.SEMICOLON) {
		p.Expression()
	}
	p.Match(lexer.SEMICOLON)
}

// declaration ";"
func (p *Parser) DeclarationStatement() {
	fmt.Println("STATEMENT-DECLARATION")
	p.Declaration()
	p.Match(lexer.SEMICOLON)
}

// ident "=" expression ";" | ident "++" ";" | ident "--" ";"
func (p *Parser) IdentStatement() {
	if p.CheckPeek(lexer.EQ) {
		fmt.Println("STATEMENT-ASSIGN")
		// Check variable exists.
		if !p.Symbols[p.CurToken.Text] {
			p.Abort("Referencing variable before assignment: " + p.CurToken.Text)
		}
		p.Match(lexer.IDENT)
		p.Match(lexer.EQ)
		p.Expression()
		p.Match(lexer.SEMICOLON)
	} else if p.CheckPeek(lexer.PLUSPLUS) {
		fmt.Println("STATEMENT-INCREMENT")
		if !p.Symbols[p.CurToken.Text] {
			p.Abort("Referencing variable before assignment: " + p.CurToken.Text)
		}
		p.Match(lexer.IDENT)
		p.Match(lexer.PLUSPLUS)
		p.Match(lexer.SEMICOLON)
	} else if p.CheckPeek(lexer.MINUSMINUS) {
		fmt.Println("STATEMENT-DECREMENT")
		if !p.Symbols[p.CurToken.Text] {
			p.Abort("Referencing variable before assignment: " + p.CurToken.Text)
		}
		p.Match(lexer.IDENT)
		p.Match(lexer.MINUSMINUS)
		p.Match(lexer.SEMICOLON)
	} else if p.CheckPeek(lexer.LPAREN) {
		// Function call: ident "(" [args] ")" ";"
		fmt.Println("STATEMENT-CALL")
		p.Match(lexer.IDENT)
		p.Match(lexer.LPAREN)
		if !p.CheckToken(lexer.RPAREN) {
			p.Args()
		}
		p.Match(lexer.RPAREN)
		p.Match(lexer.SEMICOLON)
	} else {
		p.Abort("Invalid statement: " + p.CurToken.Text)
	}
}

// declaration ::= type ident ["=" expression]
func (p *Parser) Declaration() {
	p.Type()
	// Add variable to symbols.
	if p.Symbols[p.CurToken.Text] {
		p.Abort("Variable already declared: " + p.CurToken.Text)
	}
	p.Symbols[p.CurToken.Text] = true
	p.Match(lexer.IDENT)
	if p.CheckToken(lexer.EQ) {
		p.NextToken()
		p.Expression()
	}
}

// Return true if the current token is a comparison operator.
func (p *Parser) IsComparisonOperator() bool {
	return p.CheckToken(lexer.GT) || p.CheckToken(lexer.GTEQ) ||
		p.CheckToken(lexer.LT) || p.CheckToken(lexer.LTEQ) ||
		p.CheckToken(lexer.EQEQ) || p.CheckToken(lexer.NOTEQ)
}

// comparison ::= expression (("==" | "!=" | ">" | ">=" | "<" | "<=") expression)+
func (p *Parser) Comparison() {
	fmt.Println("COMPARISON")
	p.Expression()
	if p.IsComparisonOperator() {
		p.NextToken()
		p.Expression()
	} else {
		p.Abort("Expected comparison operator at: " + p.CurToken.Text)
	}
	for p.IsComparisonOperator() {
		p.NextToken()
		p.Expression()
	}
}

// expression ::= term {("+" | "-") term}
func (p *Parser) Expression() {
	p.Term()
	for p.CheckToken(lexer.PLUS) || p.CheckToken(lexer.MINUS) {
		p.NextToken()
		p.Term()
	}
}

// term ::= unary {("/" | "*" | "%") unary}
func (p *Parser) Term() {
	p.Unary()
	for p.CheckToken(lexer.ASTERISK) || p.CheckToken(lexer.SLASH) || p.CheckToken(lexer.PERCENT) {
		p.NextToken()
		p.Unary()
	}
}

// unary ::= ["+" | "-" | "!"] primary
func (p *Parser) Unary() {
	if p.CheckToken(lexer.PLUS) || p.CheckToken(lexer.MINUS) || p.CheckToken(lexer.NOT) {
		p.NextToken()
	}
	p.Primary()
}

// primary ::= number | ident | string | "(" expression ")" | ident "(" [args] ")"
func (p *Parser) Primary() {
	if p.CheckToken(lexer.NUMBER) {
		p.NextToken()
	} else if p.CheckToken(lexer.STRING) {
		p.NextToken()
	} else if p.CheckToken(lexer.IDENT) {
		if p.CheckPeek(lexer.LPAREN) {
			// Function call: ident "(" [args] ")"
			p.NextToken()
			p.Match(lexer.LPAREN)
			if !p.CheckToken(lexer.RPAREN) {
				p.Args()
			}
			p.Match(lexer.RPAREN)
		} else {
			// Variable reference.
			if !p.Symbols[p.CurToken.Text] {
				p.Abort("Referencing variable before assignment: " + p.CurToken.Text)
			}
			p.NextToken()
		}
	} else if p.CheckToken(lexer.LPAREN) {
		// Grouped expression: "(" expression ")"
		p.NextToken()
		p.Expression()
		p.Match(lexer.RPAREN)
	} else {
		p.Abort("Unexpected token at " + p.CurToken.Text)
	}
}

// args ::= expression {"," expression}
func (p *Parser) Args() {
	p.Expression()
	for p.CheckToken(lexer.COMMA) {
		p.NextToken()
		p.Expression()
	}
}
