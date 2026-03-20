package lexer

type TokenType int

const (
	EOF     TokenType = -1
	NEWLINE TokenType = 0
	NUMBER  TokenType = 1
	IDENT   TokenType = 2
	STRING  TokenType = 3

	// Keywords
	LABEL    TokenType = 101
	GOTO     TokenType = 102
	PRINT    TokenType = 103
	INPUT    TokenType = 104
	LET      TokenType = 105
	IF       TokenType = 106
	THEN     TokenType = 107
	ENDIF    TokenType = 108
	WHILE    TokenType = 109
	REPEAT   TokenType = 110
	ENDWHILE TokenType = 111

	// Operators
	EQ       TokenType = 201
	PLUS     TokenType = 202
	MINUS    TokenType = 203
	ASTERISK TokenType = 204
	SLASH    TokenType = 205
	EQEQ     TokenType = 206
	NOTEQ    TokenType = 207
	LT       TokenType = 208
	LTEQ     TokenType = 209
	GT       TokenType = 210
	GTEQ     TokenType = 211
)

type Token struct {
	Text string
	Kind TokenType
}

func NewToken(text string, kind TokenType) Token {
	return Token{Text: text, Kind: kind}
}

func CheckIfKeyword(text string) TokenType {
	keywords := map[string]TokenType{
		"LABEL":    LABEL,
		"GOTO":     GOTO,
		"PRINT":    PRINT,
		"INPUT":    INPUT,
		"LET":      LET,
		"IF":       IF,
		"THEN":     THEN,
		"ENDIF":    ENDIF,
		"WHILE":    WHILE,
		"REPEAT":   REPEAT,
		"ENDWHILE": ENDWHILE,
	}
	if kind, ok := keywords[text]; ok {
		return kind
	}
	return -1
}
