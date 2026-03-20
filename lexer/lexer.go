package lexer

import (
	"fmt"
	"os"
)

type Lexer struct {
	Source  string
	CurChar byte
	CurPos  int
}

func NewLexer(source string) *Lexer {
	l := &Lexer{
		Source: source + "\n",
		CurPos: -1,
	}
	l.NextChar()
	return l
}

// Process the next character
func (l *Lexer) NextChar() {
	l.CurPos++
	if l.CurPos >= len(l.Source) {
		l.CurChar = 0 // EOF
	} else {
		l.CurChar = l.Source[l.CurPos]
	}
}

// Return the lookahead character
func (l *Lexer) Peek() byte {
	if l.CurPos+1 >= len(l.Source) {
		return 0
	}
	return l.Source[l.CurPos+1]
}

// Invalid token found, print error message and exit
func (l *Lexer) Abort(message string) {
	fmt.Println("Lexing Error: " + message)
	os.Exit(1)
}

// Skip whitespace including newlines (C doesn't care about newlines)
func (l *Lexer) SkipWhitespace() {
	for l.CurChar == ' ' || l.CurChar == '\t' || l.CurChar == '\r' || l.CurChar == '\n' {
		l.NextChar()
	}
}

// Skip // line comments and /* */ block comments
func (l *Lexer) SkipComments() {
	if l.CurChar == '/' && l.Peek() == '/' {
		// Line comment: skip until newline
		for l.CurChar != '\n' {
			l.NextChar()
		}
		l.NextChar() // skip the newline too
	} else if l.CurChar == '/' && l.Peek() == '*' {
		// Block comment: skip until */
		l.NextChar() // skip /
		l.NextChar() // skip *
		for !(l.CurChar == '*' && l.Peek() == '/') {
			if l.CurChar == 0 {
				l.Abort("Unterminated block comment.")
			}
			l.NextChar()
		}
		l.NextChar() // skip *
		l.NextChar() // skip /
	}
}

// Return the next token
func (l *Lexer) GetToken() Token {
	l.SkipWhitespace()
	l.SkipComments()
	l.SkipWhitespace()
	var token Token

	switch l.CurChar {
	// Delimiters
	case '{':
		token = NewToken(string(l.CurChar), LBRACE)
	case '}':
		token = NewToken(string(l.CurChar), RBRACE)
	case '(':
		token = NewToken(string(l.CurChar), LPAREN)
	case ')':
		token = NewToken(string(l.CurChar), RPAREN)
	case '[':
		token = NewToken(string(l.CurChar), LBRACKET)
	case ']':
		token = NewToken(string(l.CurChar), RBRACKET)
	case ';':
		token = NewToken(string(l.CurChar), SEMICOLON)
	case ',':
		token = NewToken(string(l.CurChar), COMMA)

	// Operators
	case '*':
		token = NewToken(string(l.CurChar), ASTERISK)
	case '%':
		token = NewToken(string(l.CurChar), PERCENT)
	case '+':
		if l.Peek() == '+' {
			lastChar := l.CurChar
			l.NextChar()
			token = NewToken(string(lastChar)+string(l.CurChar), PLUSPLUS)
		} else { 
			token = NewToken(string(l.CurChar), PLUS)
		}
	case '-':
		if l.Peek() == '-' {
			lastChar := l.CurChar
			l.NextChar()
			token = NewToken(string(lastChar)+string(l.CurChar), MINUSMINUS)
		} else {
			token = NewToken(string(l.CurChar), MINUS)
		}
	case '/':
		token = NewToken(string(l.CurChar), SLASH)
	case '=':
		if l.Peek() == '=' {
			lastChar := l.CurChar
			l.NextChar()
			token = NewToken(string(lastChar)+string(l.CurChar), EQEQ)
		} else {
			token = NewToken(string(l.CurChar), EQ)
		}
	case '>':
		if l.Peek() == '=' {
			lastChar := l.CurChar
			l.NextChar()
			token = NewToken(string(lastChar)+string(l.CurChar), GTEQ)
		} else {
			token = NewToken(string(l.CurChar), GT)
		}
	case '<':
		if l.Peek() == '=' {
			lastChar := l.CurChar
			l.NextChar()
			token = NewToken(string(lastChar)+string(l.CurChar), LTEQ)
		} else {
			token = NewToken(string(l.CurChar), LT)
		}
	case '!':
		if l.Peek() == '=' {
			lastChar := l.CurChar
			l.NextChar()
			token = NewToken(string(lastChar)+string(l.CurChar), NOTEQ)
		} else {
			token = NewToken(string(l.CurChar), NOT)
		}
	case '&':
		if l.Peek() == '&' {
			lastChar := l.CurChar
			l.NextChar()
			token = NewToken(string(lastChar)+string(l.CurChar), AND)
		} else {
			l.Abort("Expected &&, got &")
		}
	case '|':
		if l.Peek() == '|' {
			lastChar := l.CurChar
			l.NextChar()
			token = NewToken(string(lastChar)+string(l.CurChar), OR)
		} else {
			l.Abort("Expected ||, got |")
		}

	case 0:
		token = NewToken("", EOF)

	default:
		if l.CurChar == '"' {
			// Get characters between quotations.
			l.NextChar()
			startPos := l.CurPos
			for l.CurChar != '"' {
				if l.CurChar == 0 {
					l.Abort("Unterminated string.")
				}
				if l.CurChar == '\\' {
					l.NextChar() // skip the escaped character
				}
				l.NextChar()
			}
			token = NewToken(l.Source[startPos:l.CurPos], STRING)

		} else if l.CurChar >= '0' && l.CurChar <= '9' {
			// Get all consecutive digits and decimal if there is one.
			startPos := l.CurPos
			for l.Peek() >= '0' && l.Peek() <= '9' {
				l.NextChar()
			}
			if l.Peek() == '.' {
				l.NextChar()
				if l.Peek() < '0' || l.Peek() > '9' {
					l.Abort("Illegal character in number.")
				}
				for l.Peek() >= '0' && l.Peek() <= '9' {
					l.NextChar()
				}
			}
			token = NewToken(l.Source[startPos:l.CurPos+1], NUMBER)

		} else if (l.CurChar >= 'a' && l.CurChar <= 'z') || (l.CurChar >= 'A' && l.CurChar <= 'Z') || l.CurChar == '_' {
			// Identifiers and keywords. C allows underscores in identifiers.
			startPos := l.CurPos
			for (l.Peek() >= 'a' && l.Peek() <= 'z') || (l.Peek() >= 'A' && l.Peek() <= 'Z') || (l.Peek() >= '0' && l.Peek() <= '9') || l.Peek() == '_' {
				l.NextChar()
			}
			tokText := l.Source[startPos : l.CurPos+1]
			keyword := CheckIfKeyword(tokText)
			if keyword == -1 {
				token = NewToken(tokText, IDENT)
			} else {
				token = NewToken(tokText, keyword)
			}

		} else {
			l.Abort("Unknown token: " + string(l.CurChar))
		}
	}
	l.NextChar()
	return token
}
