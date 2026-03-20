package lexer

import (
	"fmt"
	"os"
)

type Lexer struct{
	Source string
	CurChar byte
	CurPos int 
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
func (l *Lexer) NextChar(){
	l.CurPos++
	if l.CurPos >= len(l.Source){
		l.CurChar = 0 // EOF 
	} else {
		l.CurChar = l.Source[l.CurPos]
	}
}

// Return the lookahead character
func (l *Lexer) Peek() byte{
	if l.CurPos + 1 >= len(l.Source){
		return 0
	}
	return l.Source[l.CurPos + 1]
}

// Invalid token found, print error message and exit 
func (l *Lexer) Abort(message string){
	fmt.Println("Lexing Error: " + message)
	os.Exit(1)
}

// Skip whitespace except newlines, which we will use to indicate the end of a statement
func (l *Lexer) SkipWhitespace(){
	for l.CurChar == ' ' || l.CurChar == '\t' || l.CurChar == '\r' {
		l.NextChar()
	}
}

// Skip commnets in the code
func (l* Lexer) SkipComments(){
	if l.CurChar == '#' {
		for l.CurChar != '\n' {
			l.NextChar()
		}
	}
}

// Return the next token
func (l *Lexer) GetToken() Token{
	l.SkipWhitespace()
	l.SkipComments()
	var token Token

	switch l.CurChar {
		// Single Character Operators
	case '+':
		token = NewToken(string(l.CurChar), PLUS)
	case '-':
		token = NewToken(string(l.CurChar), MINUS)
	case '*':
		token = NewToken(string(l.CurChar), ASTERISK)
	case '/':
		token = NewToken(string(l.CurChar), SLASH)
	case '\n':
		token = NewToken(string(l.CurChar), NEWLINE)
	case 0:
		token = NewToken("", EOF)
		// Double Character Operators
	case '=':
		if l.Peek() == '='{
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
		if l.Peek() == '='{
			lastChar := l.CurChar
			l.NextChar()
			token = NewToken(string(lastChar)+string(l.CurChar), LTEQ)
		} else {
			token = NewToken(string(l.CurChar), LT)
		}
	case '!':
		if l.Peek() == '='{
			lastChar := l.CurChar
			l.NextChar()
			token = NewToken(string(lastChar)+string(l.CurChar), NOTEQ)
		} else {
			l.Abort("Expected !=, got !" + string(l.Peek()))
		}

	default:
		if l.CurChar == '"' {
			// Get characters between quotations.
			l.NextChar()
			startPos := l.CurPos
			for l.CurChar != '"' {
				// Don't allow special characters in the string.
				if l.CurChar == '\r' || l.CurChar == '\n' || l.CurChar == '\t' || l.CurChar == '\\' || l.CurChar == '%' {
					l.Abort("Illegal character in string.")
				}
				l.NextChar()
			}
			token = NewToken(l.Source[startPos:l.CurPos], STRING)

		} else if l.CurChar >= '0' && l.CurChar <= '9' {
			// Leading character is a digit, so this must be a number.
			// Get all consecutive digits and decimal if there is one.
			startPos := l.CurPos
			for l.Peek() >= '0' && l.Peek() <= '9' {
				l.NextChar()
			}
			if l.Peek() == '.' { // Decimal!
				l.NextChar()
				// Must have at least one digit after decimal.
				if l.Peek() < '0' || l.Peek() > '9' {
					l.Abort("Illegal character in number.")
				}
				for l.Peek() >= '0' && l.Peek() <= '9' {
					l.NextChar()
				}
			}
			token = NewToken(l.Source[startPos:l.CurPos+1], NUMBER)

		} else if (l.CurChar >= 'a' && l.CurChar <= 'z') || (l.CurChar >= 'A' && l.CurChar <= 'Z') {
			// Leading character is a letter, so this must be an identifier or a keyword.
			// Get all consecutive alpha numeric characters.
			startPos := l.CurPos
			for (l.Peek() >= 'a' && l.Peek() <= 'z') || (l.Peek() >= 'A' && l.Peek() <= 'Z') || (l.Peek() >= '0' && l.Peek() <= '9') {
				l.NextChar()
			}
			// Check if the token is in the list of keywords.
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


