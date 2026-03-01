package thrift

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Document represents a parsed Thrift IDL file.
type Document struct {
	Filename   string
	Namespace  string
	Includes   []string
	Services   []*Service
	Structs    []*Struct
	Enums      []*Enum
	Exceptions []*Struct // same shape as structs
	Typedefs   []*Typedef
}

type Service struct {
	Name    string
	Methods []*Method
}

type Method struct {
	Name       string
	ReturnType *FieldType
	Params     []*Field
	Throws     []*Field
}

type Struct struct {
	Name   string
	Fields []*Field
}

type Field struct {
	ID       int
	Name     string
	Type     *FieldType
	Required bool
	Optional bool
}

type FieldType struct {
	Name     string // base type name or struct reference
	KeyType  *FieldType // for map
	ValType  *FieldType // for map, list, set
}

func (ft *FieldType) String() string {
	switch ft.Name {
	case "list":
		return fmt.Sprintf("list<%s>", ft.ValType.String())
	case "set":
		return fmt.Sprintf("set<%s>", ft.ValType.String())
	case "map":
		return fmt.Sprintf("map<%s,%s>", ft.KeyType.String(), ft.ValType.String())
	default:
		return ft.Name
	}
}

type Enum struct {
	Name   string
	Values []*EnumValue
}

type EnumValue struct {
	Name  string
	Value int
}

type Typedef struct {
	Name string
	Type *FieldType
}

// Parser parses Thrift IDL tokens into a Document.
type Parser struct {
	tokens  []Token
	pos     int
	idlRoot string
	parsed  map[string]*Document // cache of already-parsed includes
}

// ParseFile parses a Thrift IDL file and resolves includes.
func ParseFile(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	p := &Parser{
		idlRoot: filepath.Dir(path),
		parsed:  make(map[string]*Document),
	}
	return p.parse(filepath.Base(path), string(data))
}

// ParseDir parses all .thrift files in a directory.
func ParseDir(dir string) ([]*Document, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s: %w", dir, err)
	}

	p := &Parser{
		idlRoot: dir,
		parsed:  make(map[string]*Document),
	}

	var docs []*Document
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".thrift" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		doc, err := p.parse(entry.Name(), string(data))
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", entry.Name(), err)
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

func (p *Parser) parse(filename, input string) (*Document, error) {
	if doc, ok := p.parsed[filename]; ok {
		return doc, nil
	}

	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, fmt.Errorf("lexing %s: %w", filename, err)
	}

	p.tokens = tokens
	p.pos = 0

	doc := &Document{Filename: filename}
	p.parsed[filename] = doc

	for !p.atEnd() {
		switch p.current().Type {
		case TokenNamespace:
			ns, err := p.parseNamespace()
			if err != nil {
				return nil, err
			}
			doc.Namespace = ns
		case TokenInclude:
			inc, err := p.parseInclude()
			if err != nil {
				return nil, err
			}
			doc.Includes = append(doc.Includes, inc)
		case TokenService:
			svc, err := p.parseService()
			if err != nil {
				return nil, err
			}
			doc.Services = append(doc.Services, svc)
		case TokenStruct:
			s, err := p.parseStruct()
			if err != nil {
				return nil, err
			}
			doc.Structs = append(doc.Structs, s)
		case TokenEnum:
			e, err := p.parseEnum()
			if err != nil {
				return nil, err
			}
			doc.Enums = append(doc.Enums, e)
		case TokenException:
			ex, err := p.parseException()
			if err != nil {
				return nil, err
			}
			doc.Exceptions = append(doc.Exceptions, ex)
		case TokenTypedef:
			td, err := p.parseTypedef()
			if err != nil {
				return nil, err
			}
			doc.Typedefs = append(doc.Typedefs, td)
		case TokenEOF:
			return doc, nil
		default:
			return nil, fmt.Errorf("%s: unexpected token %q at line %d", filename, p.current().Value, p.current().Line)
		}
	}

	return doc, nil
}

func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() Token {
	tok := p.current()
	p.pos++
	return tok
}

func (p *Parser) expect(tt TokenType) (Token, error) {
	tok := p.current()
	if tok.Type != tt {
		return tok, fmt.Errorf("expected token type %d, got %q (%d) at line %d", tt, tok.Value, tok.Type, tok.Line)
	}
	p.pos++
	return tok, nil
}

func (p *Parser) atEnd() bool {
	return p.pos >= len(p.tokens) || p.tokens[p.pos].Type == TokenEOF
}

func (p *Parser) skipSeparator() {
	if !p.atEnd() && (p.current().Type == TokenSemicolon || p.current().Type == TokenComma) {
		p.advance()
	}
}

func (p *Parser) parseNamespace() (string, error) {
	p.advance() // skip 'namespace'
	p.advance() // skip language (e.g., 'go')
	tok, err := p.expect(TokenIdent)
	if err != nil {
		return "", err
	}
	return tok.Value, nil
}

func (p *Parser) parseInclude() (string, error) {
	p.advance() // skip 'include'
	tok, err := p.expect(TokenString)
	if err != nil {
		return "", err
	}
	return tok.Value, nil
}

func (p *Parser) parseService() (*Service, error) {
	p.advance() // skip 'service'
	nameTok, err := p.expect(TokenIdent)
	if err != nil {
		return nil, err
	}

	svc := &Service{Name: nameTok.Value}

	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	for p.current().Type != TokenRBrace && !p.atEnd() {
		method, err := p.parseMethod()
		if err != nil {
			return nil, err
		}
		svc.Methods = append(svc.Methods, method)
		p.skipSeparator()
	}

	if _, err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}

	return svc, nil
}

func (p *Parser) parseMethod() (*Method, error) {
	retType, err := p.parseFieldType()
	if err != nil {
		return nil, err
	}

	nameTok, err := p.expect(TokenIdent)
	if err != nil {
		return nil, err
	}

	method := &Method{
		Name:       nameTok.Value,
		ReturnType: retType,
	}

	// Parse params
	if _, err := p.expect(TokenLParen); err != nil {
		return nil, err
	}
	for p.current().Type != TokenRParen && !p.atEnd() {
		field, err := p.parseField()
		if err != nil {
			return nil, err
		}
		method.Params = append(method.Params, field)
		p.skipSeparator()
	}
	if _, err := p.expect(TokenRParen); err != nil {
		return nil, err
	}

	// Parse throws
	if p.current().Type == TokenThrows {
		p.advance()
		if _, err := p.expect(TokenLParen); err != nil {
			return nil, err
		}
		for p.current().Type != TokenRParen && !p.atEnd() {
			field, err := p.parseField()
			if err != nil {
				return nil, err
			}
			method.Throws = append(method.Throws, field)
			p.skipSeparator()
		}
		if _, err := p.expect(TokenRParen); err != nil {
			return nil, err
		}
	}

	p.skipSeparator()
	return method, nil
}

func (p *Parser) parseStruct() (*Struct, error) {
	p.advance() // skip 'struct'
	nameTok, err := p.expect(TokenIdent)
	if err != nil {
		return nil, err
	}

	s := &Struct{Name: nameTok.Value}

	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	for p.current().Type != TokenRBrace && !p.atEnd() {
		field, err := p.parseField()
		if err != nil {
			return nil, err
		}
		s.Fields = append(s.Fields, field)
		p.skipSeparator()
	}

	if _, err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}

	return s, nil
}

func (p *Parser) parseException() (*Struct, error) {
	p.advance() // skip 'exception'
	nameTok, err := p.expect(TokenIdent)
	if err != nil {
		return nil, err
	}

	s := &Struct{Name: nameTok.Value}

	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	for p.current().Type != TokenRBrace && !p.atEnd() {
		field, err := p.parseField()
		if err != nil {
			return nil, err
		}
		s.Fields = append(s.Fields, field)
		p.skipSeparator()
	}

	if _, err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}

	return s, nil
}

func (p *Parser) parseEnum() (*Enum, error) {
	p.advance() // skip 'enum'
	nameTok, err := p.expect(TokenIdent)
	if err != nil {
		return nil, err
	}

	e := &Enum{Name: nameTok.Value}

	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	for p.current().Type != TokenRBrace && !p.atEnd() {
		valName, err := p.expect(TokenIdent)
		if err != nil {
			return nil, err
		}
		ev := &EnumValue{Name: valName.Value}
		if p.current().Type == TokenEquals {
			p.advance()
			numTok, err := p.expect(TokenNumber)
			if err != nil {
				return nil, err
			}
			ev.Value, _ = strconv.Atoi(numTok.Value)
		}
		e.Values = append(e.Values, ev)
		p.skipSeparator()
	}

	if _, err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}

	return e, nil
}

func (p *Parser) parseTypedef() (*Typedef, error) {
	p.advance() // skip 'typedef'
	ft, err := p.parseFieldType()
	if err != nil {
		return nil, err
	}
	nameTok, err := p.expect(TokenIdent)
	if err != nil {
		return nil, err
	}
	return &Typedef{Name: nameTok.Value, Type: ft}, nil
}

func (p *Parser) parseField() (*Field, error) {
	field := &Field{}

	// Field ID (optional)
	if p.current().Type == TokenNumber {
		numTok := p.advance()
		field.ID, _ = strconv.Atoi(numTok.Value)
		if _, err := p.expect(TokenColon); err != nil {
			return nil, err
		}
	}

	// Required/Optional
	if p.current().Type == TokenRequired {
		field.Required = true
		p.advance()
	} else if p.current().Type == TokenOptional {
		field.Optional = true
		p.advance()
	}

	// Field type
	ft, err := p.parseFieldType()
	if err != nil {
		return nil, err
	}
	field.Type = ft

	// Field name
	nameTok, err := p.expect(TokenIdent)
	if err != nil {
		return nil, err
	}
	field.Name = nameTok.Value

	return field, nil
}

func (p *Parser) parseFieldType() (*FieldType, error) {
	tok := p.current()

	switch tok.Type {
	case TokenList:
		p.advance()
		if _, err := p.expect(TokenLAngle); err != nil {
			return nil, err
		}
		valType, err := p.parseFieldType()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TokenRAngle); err != nil {
			return nil, err
		}
		return &FieldType{Name: "list", ValType: valType}, nil

	case TokenSet:
		p.advance()
		if _, err := p.expect(TokenLAngle); err != nil {
			return nil, err
		}
		valType, err := p.parseFieldType()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TokenRAngle); err != nil {
			return nil, err
		}
		return &FieldType{Name: "set", ValType: valType}, nil

	case TokenMap:
		p.advance()
		if _, err := p.expect(TokenLAngle); err != nil {
			return nil, err
		}
		keyType, err := p.parseFieldType()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TokenComma); err != nil {
			return nil, err
		}
		valType, err := p.parseFieldType()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TokenRAngle); err != nil {
			return nil, err
		}
		return &FieldType{Name: "map", KeyType: keyType, ValType: valType}, nil

	case TokenVoid:
		p.advance()
		return &FieldType{Name: "void"}, nil
	case TokenBool:
		p.advance()
		return &FieldType{Name: "bool"}, nil
	case TokenByte:
		p.advance()
		return &FieldType{Name: "byte"}, nil
	case TokenI16:
		p.advance()
		return &FieldType{Name: "i16"}, nil
	case TokenI32:
		p.advance()
		return &FieldType{Name: "i32"}, nil
	case TokenI64:
		p.advance()
		return &FieldType{Name: "i64"}, nil
	case TokenDouble:
		p.advance()
		return &FieldType{Name: "double"}, nil
	case TokenString_:
		p.advance()
		return &FieldType{Name: "string"}, nil
	case TokenBinary:
		p.advance()
		return &FieldType{Name: "binary"}, nil

	case TokenIdent:
		p.advance()
		name := tok.Value
		// Handle qualified names like "common.BaseResponse"
		if p.current().Type == TokenDot {
			p.advance()
			qualTok, err := p.expect(TokenIdent)
			if err != nil {
				return nil, err
			}
			name = name + "." + qualTok.Value
		}
		return &FieldType{Name: name}, nil

	default:
		return nil, fmt.Errorf("expected type, got %q at line %d", tok.Value, tok.Line)
	}
}
