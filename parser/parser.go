package parser

import (
	"C_Compiler/emitter"
	"C_Compiler/lexer"
	"fmt"
	"os"
)

type Parser struct {
	Lexer     *lexer.Lexer
	Emitter   *emitter.Emitter
	CurToken  lexer.Token
	PeekToken lexer.Token
	Symbols   map[string]bool // Variables declared so far.
}

func NewParser(l *lexer.Lexer, e *emitter.Emitter) *Parser {
	p := &Parser{
		Lexer:   l,
		Emitter: e,
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

// Emit the inverted conditional jump for a comparison operator.
func (p *Parser) EmitConditionalJump(op lexer.TokenType, label string) {
	switch op {
	case lexer.EQEQ:
		p.Emitter.EmitLine("jne " + label)
	case lexer.NOTEQ:
		p.Emitter.EmitLine("je " + label)
	case lexer.GT:
		p.Emitter.EmitLine("jle " + label)
	case lexer.GTEQ:
		p.Emitter.EmitLine("jl " + label)
	case lexer.LT:
		p.Emitter.EmitLine("jge " + label)
	case lexer.LTEQ:
		p.Emitter.EmitLine("jg " + label)
	}
}

// program ::= {function | declaration}
func (p *Parser) Program() {
	p.Emitter.Emit("global main\n")
	for !p.CheckToken(lexer.EOF) {
		p.Function()
	}
}

// function ::= type ident "(" [params] ")" block
func (p *Parser) Function() {
	p.Type()

	// Save function name before Match consumes it
	funcName := p.CurToken.Text
	p.Match(lexer.IDENT)

	// Emit function label and prologue
	p.Emitter.EmitLabel(funcName)
	p.Emitter.EmitLine("push rbp")
	p.Emitter.EmitLine("mov rbp, rsp")

	// Save position — we'll insert "sub rsp, N" here later
	subRspPos := len(p.Emitter.Code)

	// Reset stack for this function
	p.Emitter.StackOffset = 0
	p.Emitter.Variables = make(map[string]int)
	p.Symbols = make(map[string]bool)

	p.Match(lexer.LPAREN)
	p.Params()
	p.Match(lexer.RPAREN)
	p.Block()

	// Now we know how much stack space we need — align to 16 bytes
	stackSize := -p.Emitter.StackOffset
	if stackSize%16 != 0 {
		stackSize += 16 - (stackSize % 16)
	}
	if stackSize > 0 {
		subLine := fmt.Sprintf("    sub rsp, %d\n", stackSize)
		p.Emitter.Code = p.Emitter.Code[:subRspPos] + subLine + p.Emitter.Code[subRspPos:]
	}

	// Emit epilogue (safety net if function doesn't end with return)
	p.Emitter.EmitLine("leave")
	p.Emitter.EmitLine("ret")
}

// params ::= type ident {"," type ident}
func (p *Parser) Params() {
	// System V ABI: first 6 args in these registers (32-bit)
	argRegs := []string{"edi", "esi", "edx", "ecx", "r8d", "r9d"}
	argCount := 0

	if !p.CheckToken(lexer.RPAREN) {
		p.Type()

		// Declare param on stack and store from register
		paramName := p.CurToken.Text
		p.Symbols[paramName] = true
		offset := p.Emitter.DeclareVariable(paramName)
		p.Emitter.EmitLine(fmt.Sprintf("mov [rbp%d], %s", offset, argRegs[argCount]))
		argCount++
		p.Match(lexer.IDENT)

		for p.CheckToken(lexer.COMMA) {
			p.NextToken()
			p.Type()

			paramName = p.CurToken.Text
			p.Symbols[paramName] = true
			offset = p.Emitter.DeclareVariable(paramName)
			p.Emitter.EmitLine(fmt.Sprintf("mov [rbp%d], %s", offset, argRegs[argCount]))
			argCount++
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
	labelNum := p.Emitter.NextLabel()
	elseLabel := fmt.Sprintf(".Lelse_%d", labelNum)
	endLabel := fmt.Sprintf(".Lendif_%d", labelNum)

	p.Match(lexer.IF)
	p.Match(lexer.LPAREN)
	op := p.Comparison()
	p.Match(lexer.RPAREN)

	// Jump PAST the if-body when condition is false (inverted jump)
	p.EmitConditionalJump(op, elseLabel)

	p.Block()

	if p.CheckToken(lexer.ELSE) {
		p.Emitter.EmitLine("jmp " + endLabel) // skip else body
		p.Emitter.EmitLabel(elseLabel)
		p.NextToken()
		p.Block()
		p.Emitter.EmitLabel(endLabel)
	} else {
		p.Emitter.EmitLabel(elseLabel)
	}
}

// "while" "(" comparison ")" block
func (p *Parser) WhileStatement() {
	labelNum := p.Emitter.NextLabel()
	startLabel := fmt.Sprintf(".Lwhile_%d", labelNum)
	endLabel := fmt.Sprintf(".Lendwhile_%d", labelNum)

	p.Emitter.EmitLabel(startLabel)

	p.Match(lexer.WHILE)
	p.Match(lexer.LPAREN)
	op := p.Comparison()
	p.Match(lexer.RPAREN)

	// Jump past body when false
	p.EmitConditionalJump(op, endLabel)

	p.Block()

	p.Emitter.EmitLine("jmp " + startLabel) // loop back
	p.Emitter.EmitLabel(endLabel)
}

// "for" "(" declaration ";" comparison ";" ident "=" expression ")" block
func (p *Parser) ForStatement() {
	labelNum := p.Emitter.NextLabel()
	startLabel := fmt.Sprintf(".Lfor_%d", labelNum)
	endLabel := fmt.Sprintf(".Lendfor_%d", labelNum)

	p.Match(lexer.FOR)
	p.Match(lexer.LPAREN)

	// Init: declaration
	p.Declaration()
	p.Match(lexer.SEMICOLON)

	p.Emitter.EmitLabel(startLabel)

	// Condition
	op := p.Comparison()
	p.Match(lexer.SEMICOLON)

	// Jump past body when false
	p.EmitConditionalJump(op, endLabel)

	// Parse update expression now but emit it after the block
	updateStart := len(p.Emitter.Code)

	varName := p.CurToken.Text
	p.Match(lexer.IDENT)
	p.Match(lexer.EQ)
	p.Expression() // result in eax
	offset := p.Emitter.GetVariable(varName)
	p.Emitter.EmitLine(fmt.Sprintf("mov [rbp%d], eax", offset))

	// Grab the update code and remove it
	updateCode := p.Emitter.Code[updateStart:]
	p.Emitter.Code = p.Emitter.Code[:updateStart]

	p.Match(lexer.RPAREN)

	p.Block()

	// Emit the update code after the body
	p.Emitter.Code += updateCode
	p.Emitter.EmitLine("jmp " + startLabel)
	p.Emitter.EmitLabel(endLabel)
}

// "return" [expression] ";"
func (p *Parser) ReturnStatement() {
	p.Match(lexer.RETURN)
	if !p.CheckToken(lexer.SEMICOLON) {
		p.Expression() // result in eax
	}
	p.Match(lexer.SEMICOLON)
	p.Emitter.EmitLine("leave")
	p.Emitter.EmitLine("ret")
}

// declaration ";"
func (p *Parser) DeclarationStatement() {
	p.Declaration()
	p.Match(lexer.SEMICOLON)
}

// ident "=" expression ";" | ident "++" ";" | ident "--" ";"
func (p *Parser) IdentStatement() {
	if p.CheckPeek(lexer.EQ) {
		// Assignment: ident = expression ;
		varName := p.CurToken.Text
		if !p.Symbols[p.CurToken.Text] {
			p.Abort("Referencing variable before assignment: " + p.CurToken.Text)
		}
		p.Match(lexer.IDENT)
		p.Match(lexer.EQ)
		p.Expression() // result in eax
		offset := p.Emitter.GetVariable(varName)
		p.Emitter.EmitLine(fmt.Sprintf("mov [rbp%d], eax", offset))
		p.Match(lexer.SEMICOLON)

	} else if p.CheckPeek(lexer.PLUSPLUS) {
		varName := p.CurToken.Text
		if !p.Symbols[p.CurToken.Text] {
			p.Abort("Referencing variable before assignment: " + p.CurToken.Text)
		}
		offset := p.Emitter.GetVariable(varName)
		p.Emitter.EmitLine(fmt.Sprintf("add dword [rbp%d], 1", offset))
		p.Match(lexer.IDENT)
		p.Match(lexer.PLUSPLUS)
		p.Match(lexer.SEMICOLON)

	} else if p.CheckPeek(lexer.MINUSMINUS) {
		varName := p.CurToken.Text
		if !p.Symbols[p.CurToken.Text] {
			p.Abort("Referencing variable before assignment: " + p.CurToken.Text)
		}
		offset := p.Emitter.GetVariable(varName)
		p.Emitter.EmitLine(fmt.Sprintf("sub dword [rbp%d], 1", offset))
		p.Match(lexer.IDENT)
		p.Match(lexer.MINUSMINUS)
		p.Match(lexer.SEMICOLON)

	} else if p.CheckPeek(lexer.LPAREN) {
		// Function call: ident "(" [args] ")" ";"
		funcName := p.CurToken.Text
		p.Match(lexer.IDENT)
		p.Match(lexer.LPAREN)
		if !p.CheckToken(lexer.RPAREN) {
			p.Args()
		}
		p.Match(lexer.RPAREN)
		p.Emitter.EmitLine("call " + funcName)
		p.Match(lexer.SEMICOLON)

	} else {
		p.Abort("Invalid statement: " + p.CurToken.Text)
	}
}

// declaration ::= type ident ["=" expression]
func (p *Parser) Declaration() {
	p.Type()
	varName := p.CurToken.Text
	if p.Symbols[varName] {
		p.Abort("Variable already declared: " + varName)
	}
	p.Symbols[varName] = true
	offset := p.Emitter.DeclareVariable(varName)
	p.Match(lexer.IDENT)
	if p.CheckToken(lexer.EQ) {
		p.NextToken()
		p.Expression() // result in eax
		p.Emitter.EmitLine(fmt.Sprintf("mov [rbp%d], eax", offset))
	}
}

// Return true if the current token is a comparison operator.
func (p *Parser) IsComparisonOperator() bool {
	return p.CheckToken(lexer.GT) || p.CheckToken(lexer.GTEQ) ||
		p.CheckToken(lexer.LT) || p.CheckToken(lexer.LTEQ) ||
		p.CheckToken(lexer.EQEQ) || p.CheckToken(lexer.NOTEQ)
}

// comparison ::= expression (("==" | "!=" | ">" | ">=" | "<" | "<=") expression)+
func (p *Parser) Comparison() lexer.TokenType {
	p.Expression() // left side → eax
	p.Emitter.EmitLine("push rax") // save left side

	if !p.IsComparisonOperator() {
		p.Abort("Expected comparison operator at: " + p.CurToken.Text)
	}
	op := p.CurToken.Kind
	p.NextToken()

	p.Expression() // right side → eax
	p.Emitter.EmitLine("pop rbx") // left side into rbx
	p.Emitter.EmitLine("cmp ebx, eax") // compare left to right

	return op
}

// expression ::= term {("+" | "-") term}
func (p *Parser) Expression() {
	p.Term() // first term → eax
	for p.CheckToken(lexer.PLUS) || p.CheckToken(lexer.MINUS) {
		op := p.CurToken.Kind
		p.Emitter.EmitLine("push rax") // save left side
		p.NextToken()
		p.Term() // right side → eax
		p.Emitter.EmitLine("pop rbx") // left side into rbx
		if op == lexer.PLUS {
			p.Emitter.EmitLine("add eax, ebx")
		} else {
			// subtraction: left - right = rbx - eax
			p.Emitter.EmitLine("sub ebx, eax")
			p.Emitter.EmitLine("mov eax, ebx")
		}
	}
}

// term ::= unary {("/" | "*" | "%") unary}
func (p *Parser) Term() {
	p.Unary() // first unary → eax
	for p.CheckToken(lexer.ASTERISK) || p.CheckToken(lexer.SLASH) || p.CheckToken(lexer.PERCENT) {
		op := p.CurToken.Kind
		p.Emitter.EmitLine("push rax") // save left side
		p.NextToken()
		p.Unary() // right side → eax
		p.Emitter.EmitLine("pop rbx") // left side into rbx
		if op == lexer.ASTERISK {
			p.Emitter.EmitLine("imul eax, ebx")
		} else if op == lexer.SLASH {
			// divide: left / right = ebx / eax
			p.Emitter.EmitLine("mov ecx, eax") // save divisor
			p.Emitter.EmitLine("mov eax, ebx") // dividend into eax
			p.Emitter.EmitLine("cdq")          // sign extend into edx:eax
			p.Emitter.EmitLine("idiv ecx")     // eax = quotient
		} else {
			// modulo: same as divide but result in edx
			p.Emitter.EmitLine("mov ecx, eax")
			p.Emitter.EmitLine("mov eax, ebx")
			p.Emitter.EmitLine("cdq")
			p.Emitter.EmitLine("idiv ecx")
			p.Emitter.EmitLine("mov eax, edx") // remainder into eax
		}
	}
}

// unary ::= ["+" | "-" | "!"] primary
func (p *Parser) Unary() {
	if p.CheckToken(lexer.MINUS) {
		p.NextToken()
		p.Primary() // value → eax
		p.Emitter.EmitLine("neg eax")
	} else if p.CheckToken(lexer.NOT) {
		p.NextToken()
		p.Primary()
		p.Emitter.EmitLine("cmp eax, 0")
		p.Emitter.EmitLine("sete al")
		p.Emitter.EmitLine("movzx eax, al")
	} else if p.CheckToken(lexer.PLUS) {
		p.NextToken()
		p.Primary() // unary + does nothing
	} else {
		p.Primary()
	}
}

// primary ::= number | ident | string | "(" expression ")" | ident "(" [args] ")"
func (p *Parser) Primary() {
	if p.CheckToken(lexer.NUMBER) {
		p.Emitter.EmitLine("mov eax, " + p.CurToken.Text)
		p.NextToken()
	} else if p.CheckToken(lexer.STRING) {
		// Store string in .data section
		strLabel := fmt.Sprintf("str_%d", p.Emitter.NextLabel())
		p.Emitter.DataLine(strLabel + `: db "` + p.CurToken.Text + `", 0`)
		p.Emitter.EmitLine("lea rax, [rel " + strLabel + "]")
		p.NextToken()
	} else if p.CheckToken(lexer.IDENT) {
		if p.CheckPeek(lexer.LPAREN) {
			// Function call: ident "(" [args] ")"
			funcName := p.CurToken.Text
			p.NextToken()
			p.Match(lexer.LPAREN)
			if !p.CheckToken(lexer.RPAREN) {
				p.Args()
			}
			p.Match(lexer.RPAREN)
			p.Emitter.EmitLine("call " + funcName)
			// result is in eax
		} else {
			// Variable reference
			if !p.Symbols[p.CurToken.Text] {
				p.Abort("Referencing variable before assignment: " + p.CurToken.Text)
			}
			offset := p.Emitter.GetVariable(p.CurToken.Text)
			p.Emitter.EmitLine(fmt.Sprintf("mov eax, [rbp%d]", offset))
			p.NextToken()
		}
	} else if p.CheckToken(lexer.LPAREN) {
		// Grouped expression: "(" expression ")"
		p.NextToken()
		p.Expression() // result in eax
		p.Match(lexer.RPAREN)
	} else {
		p.Abort("Unexpected token at " + p.CurToken.Text)
	}
}

// args ::= expression {"," expression}
func (p *Parser) Args() {
	argRegs := []string{"edi", "esi", "edx", "ecx", "r8d", "r9d"}
	argCount := 0

	p.Expression() // first arg → eax
	p.Emitter.EmitLine("push rax") // save it
	argCount++

	for p.CheckToken(lexer.COMMA) {
		p.NextToken()
		p.Expression()
		p.Emitter.EmitLine("push rax")
		argCount++
	}

	// Pop args into registers in reverse order
	for i := argCount - 1; i >= 0; i-- {
		p.Emitter.EmitLine("pop rax")
		p.Emitter.EmitLine(fmt.Sprintf("mov %s, eax", argRegs[i]))
	}
}
