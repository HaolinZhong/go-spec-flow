package thrift

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdent
	TokenNumber
	TokenString
	TokenLBrace    // {
	TokenRBrace    // }
	TokenLParen    // (
	TokenRParen    // )
	TokenLAngle    // <
	TokenRAngle    // >
	TokenColon     // :
	TokenSemicolon // ;
	TokenComma     // ,
	TokenEquals    // =
	TokenDot       // .

	// Keywords
	TokenNamespace
	TokenInclude
	TokenService
	TokenStruct
	TokenEnum
	TokenException
	TokenTypedef
	TokenRequired
	TokenOptional
	TokenThrows
	TokenList
	TokenMap
	TokenSet
	TokenVoid
	TokenBool
	TokenByte
	TokenI16
	TokenI32
	TokenI64
	TokenDouble
	TokenString_
	TokenBinary
)

var keywords = map[string]TokenType{
	"namespace": TokenNamespace,
	"include":   TokenInclude,
	"service":   TokenService,
	"struct":    TokenStruct,
	"enum":      TokenEnum,
	"exception": TokenException,
	"typedef":   TokenTypedef,
	"required":  TokenRequired,
	"optional":  TokenOptional,
	"throws":    TokenThrows,
	"list":      TokenList,
	"map":       TokenMap,
	"set":       TokenSet,
	"void":      TokenVoid,
	"bool":      TokenBool,
	"byte":      TokenByte,
	"i16":       TokenI16,
	"i32":       TokenI32,
	"i64":       TokenI64,
	"double":    TokenDouble,
	"string":    TokenString_,
	"binary":    TokenBinary,
}

type Token struct {
	Type    TokenType
	Value   string
	Line    int
	Col     int
}

type Lexer struct {
	input  string
	pos    int
	line   int
	col    int
	tokens []Token
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input, line: 1, col: 1}
}

func (l *Lexer) Tokenize() ([]Token, error) {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]

		// Skip whitespace
		if unicode.IsSpace(rune(ch)) {
			if ch == '\n' {
				l.line++
				l.col = 1
			} else {
				l.col++
			}
			l.pos++
			continue
		}

		// Skip comments
		if ch == '/' && l.pos+1 < len(l.input) {
			if l.input[l.pos+1] == '/' {
				l.skipLineComment()
				continue
			}
			if l.input[l.pos+1] == '*' {
				l.skipBlockComment()
				continue
			}
		}
		if ch == '#' {
			l.skipLineComment()
			continue
		}

		// String literal
		if ch == '"' {
			tok, err := l.readString()
			if err != nil {
				return nil, err
			}
			l.tokens = append(l.tokens, tok)
			continue
		}

		// Number
		if unicode.IsDigit(rune(ch)) || (ch == '-' && l.pos+1 < len(l.input) && unicode.IsDigit(rune(l.input[l.pos+1]))) {
			l.tokens = append(l.tokens, l.readNumber())
			continue
		}

		// Identifier or keyword
		if unicode.IsLetter(rune(ch)) || ch == '_' {
			l.tokens = append(l.tokens, l.readIdent())
			continue
		}

		// Punctuation
		tok := Token{Line: l.line, Col: l.col, Value: string(ch)}
		switch ch {
		case '{':
			tok.Type = TokenLBrace
		case '}':
			tok.Type = TokenRBrace
		case '(':
			tok.Type = TokenLParen
		case ')':
			tok.Type = TokenRParen
		case '<':
			tok.Type = TokenLAngle
		case '>':
			tok.Type = TokenRAngle
		case ':':
			tok.Type = TokenColon
		case ';':
			tok.Type = TokenSemicolon
		case ',':
			tok.Type = TokenComma
		case '=':
			tok.Type = TokenEquals
		case '.':
			tok.Type = TokenDot
		default:
			return nil, fmt.Errorf("unexpected character '%c' at line %d col %d", ch, l.line, l.col)
		}
		l.tokens = append(l.tokens, tok)
		l.pos++
		l.col++
	}

	l.tokens = append(l.tokens, Token{Type: TokenEOF, Line: l.line, Col: l.col})
	return l.tokens, nil
}

func (l *Lexer) skipLineComment() {
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.pos++
	}
}

func (l *Lexer) skipBlockComment() {
	l.pos += 2
	l.col += 2
	for l.pos+1 < len(l.input) {
		if l.input[l.pos] == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
		if l.input[l.pos] == '*' && l.input[l.pos+1] == '/' {
			l.pos += 2
			l.col += 2
			return
		}
		l.pos++
	}
}

func (l *Lexer) readString() (Token, error) {
	start := l.pos
	startLine, startCol := l.line, l.col
	l.pos++ // skip opening quote
	l.col++
	var sb strings.Builder
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '"' {
			l.pos++
			l.col++
			return Token{Type: TokenString, Value: sb.String(), Line: startLine, Col: startCol}, nil
		}
		if ch == '\\' && l.pos+1 < len(l.input) {
			l.pos++
			l.col++
			ch = l.input[l.pos]
		}
		sb.WriteByte(ch)
		l.pos++
		l.col++
	}
	return Token{}, fmt.Errorf("unterminated string at line %d col %d", startLine, startCol-start)
}

func (l *Lexer) readNumber() Token {
	start := l.pos
	startCol := l.col
	if l.input[l.pos] == '-' {
		l.pos++
		l.col++
	}
	for l.pos < len(l.input) && unicode.IsDigit(rune(l.input[l.pos])) {
		l.pos++
		l.col++
	}
	return Token{Type: TokenNumber, Value: l.input[start:l.pos], Line: l.line, Col: startCol}
}

func (l *Lexer) readIdent() Token {
	start := l.pos
	startCol := l.col
	for l.pos < len(l.input) && (unicode.IsLetter(rune(l.input[l.pos])) || unicode.IsDigit(rune(l.input[l.pos])) || l.input[l.pos] == '_') {
		l.pos++
		l.col++
	}
	value := l.input[start:l.pos]
	if tt, ok := keywords[value]; ok {
		return Token{Type: tt, Value: value, Line: l.line, Col: startCol}
	}
	return Token{Type: TokenIdent, Value: value, Line: l.line, Col: startCol}
}
