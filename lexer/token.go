package lexer

type TokenType int

const (
	EOF     TokenType = -1
	NEWLINE TokenType = 0
	NUMBER  TokenType = 1
	IDENT   TokenType = 2
	STRING  TokenType = 3

	// Keywords
	INT    TokenType = 101
	FLOAT  TokenType = 102
	CHAR   TokenType = 103
	VOID   TokenType = 104
	IF     TokenType = 105
	ELSE   TokenType = 106
	WHILE  TokenType = 107
	FOR    TokenType = 108
	RETURN TokenType = 109
	STRUCT TokenType = 110

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
	PERCENT  TokenType = 212
	AND      TokenType = 213
	OR       TokenType = 214
	NOT      TokenType = 215
	PLUSPLUS  TokenType = 216
	MINUSMINUS TokenType = 217

	// Delimiters
	LBRACE    TokenType = 301
	RBRACE    TokenType = 302
	LPAREN    TokenType = 303
	RPAREN    TokenType = 304
	LBRACKET  TokenType = 305
	RBRACKET  TokenType = 306
	SEMICOLON TokenType = 307
	COMMA     TokenType = 308
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
		"int":    INT,
		"float":  FLOAT,
		"char":   CHAR,
		"void":   VOID,
		"if":     IF,
		"else":   ELSE,
		"while":  WHILE,
		"for":    FOR,
		"return": RETURN,
		"struct": STRUCT,
	}
	if kind, ok := keywords[text]; ok {
		return kind
	}
	return -1
}
