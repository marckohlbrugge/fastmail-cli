package jmap

import (
	"strings"
	"unicode"
)

// Filter represents a JMAP email filter.
type Filter interface {
	ToJMAP() map[string]interface{}
}

// TextFilter represents a simple field filter.
type TextFilter struct {
	Field string // "text", "from", "to", "subject", "inMailbox", "hasKeyword", "notKeyword"
	Value string
}

// ToJMAP converts the filter to JMAP format.
func (f *TextFilter) ToJMAP() map[string]interface{} {
	return map[string]interface{}{f.Field: f.Value}
}

// BoolFilter represents a boolean operator filter (AND, OR, NOT).
type BoolFilter struct {
	Operator   string // "AND", "OR", "NOT"
	Conditions []Filter
}

// ToJMAP converts the filter to JMAP format.
func (f *BoolFilter) ToJMAP() map[string]interface{} {
	conditions := make([]map[string]interface{}, len(f.Conditions))
	for i, c := range f.Conditions {
		conditions[i] = c.ToJMAP()
	}
	return map[string]interface{}{
		"operator":   f.Operator,
		"conditions": conditions,
	}
}

// HasAttachmentFilter represents an attachment filter.
type HasAttachmentFilter struct {
	Value bool
}

// ToJMAP converts the filter to JMAP format.
func (f *HasAttachmentFilter) ToJMAP() map[string]interface{} {
	return map[string]interface{}{"hasAttachment": f.Value}
}

// tokenType represents the type of a token.
type tokenType int

const (
	tokenWord tokenType = iota
	tokenAND
	tokenOR
	tokenNOT
	tokenLParen
	tokenRParen
	tokenEOF
)

// token represents a lexer token.
type token struct {
	typ   tokenType
	value string
}

// tokenizer splits a query string into tokens.
type tokenizer struct {
	input string
	pos   int
}

func newTokenizer(input string) *tokenizer {
	return &tokenizer{input: input, pos: 0}
}

func (t *tokenizer) peek() byte {
	if t.pos >= len(t.input) {
		return 0
	}
	return t.input[t.pos]
}

func (t *tokenizer) advance() {
	t.pos++
}

func (t *tokenizer) skipWhitespace() {
	for t.pos < len(t.input) && unicode.IsSpace(rune(t.input[t.pos])) {
		t.pos++
	}
}

func (t *tokenizer) readQuoted() string {
	quote := t.peek()
	t.advance() // skip opening quote
	start := t.pos
	for t.pos < len(t.input) && t.input[t.pos] != quote {
		if t.input[t.pos] == '\\' && t.pos+1 < len(t.input) {
			t.pos += 2 // skip escaped character
		} else {
			t.pos++
		}
	}
	value := t.input[start:t.pos]
	if t.pos < len(t.input) {
		t.advance() // skip closing quote
	}
	// Remove escape characters
	value = strings.ReplaceAll(value, "\\\"", "\"")
	value = strings.ReplaceAll(value, "\\'", "'")
	return value
}

func (t *tokenizer) readWord() string {
	start := t.pos
	for t.pos < len(t.input) {
		c := t.input[t.pos]
		if unicode.IsSpace(rune(c)) || c == '(' || c == ')' {
			break
		}
		// Handle colon followed by quoted string
		if c == ':' && t.pos+1 < len(t.input) && (t.input[t.pos+1] == '"' || t.input[t.pos+1] == '\'') {
			field := t.input[start:t.pos]
			t.pos++ // skip the colon
			quoted := t.readQuoted()
			return field + ":" + quoted
		}
		t.pos++
	}
	return t.input[start:t.pos]
}

func (t *tokenizer) nextToken() token {
	t.skipWhitespace()

	if t.pos >= len(t.input) {
		return token{typ: tokenEOF}
	}

	c := t.peek()

	switch c {
	case '(':
		t.advance()
		return token{typ: tokenLParen, value: "("}
	case ')':
		t.advance()
		return token{typ: tokenRParen, value: ")"}
	case '"', '\'':
		return token{typ: tokenWord, value: t.readQuoted()}
	}

	word := t.readWord()
	upper := strings.ToUpper(word)

	switch upper {
	case "AND":
		return token{typ: tokenAND, value: word}
	case "OR":
		return token{typ: tokenOR, value: word}
	case "NOT":
		return token{typ: tokenNOT, value: word}
	default:
		return token{typ: tokenWord, value: word}
	}
}

// parser parses a query string into a Filter.
type parser struct {
	tokenizer *tokenizer
	current   token
}

func newParser(input string) *parser {
	p := &parser{tokenizer: newTokenizer(input)}
	p.advance()
	return p
}

func (p *parser) advance() {
	p.current = p.tokenizer.nextToken()
}

// Parse parses a query string into a Filter.
// Grammar:
//
//	expr     -> orExpr
//	orExpr   -> andExpr ("OR" andExpr)*
//	andExpr  -> notExpr (("AND")? notExpr)*
//	notExpr  -> "NOT" notExpr | primary
//	primary  -> "(" expr ")" | term
//	term     -> field:value | text
func (p *parser) parse() Filter {
	if p.current.typ == tokenEOF {
		return nil
	}
	return p.parseOrExpr()
}

func (p *parser) parseOrExpr() Filter {
	left := p.parseAndExpr()
	if left == nil {
		return nil
	}

	conditions := []Filter{left}
	for p.current.typ == tokenOR {
		p.advance()
		right := p.parseAndExpr()
		if right != nil {
			conditions = append(conditions, right)
		}
	}

	if len(conditions) == 1 {
		return conditions[0]
	}
	return &BoolFilter{Operator: "OR", Conditions: conditions}
}

func (p *parser) parseAndExpr() Filter {
	left := p.parseNotExpr()
	if left == nil {
		return nil
	}

	conditions := []Filter{left}
	for {
		// Explicit AND
		if p.current.typ == tokenAND {
			p.advance()
			right := p.parseNotExpr()
			if right != nil {
				conditions = append(conditions, right)
			}
			continue
		}

		// Implicit AND: next token is a term (word, NOT, or parenthesis)
		if p.current.typ == tokenWord || p.current.typ == tokenNOT || p.current.typ == tokenLParen {
			right := p.parseNotExpr()
			if right != nil {
				conditions = append(conditions, right)
			}
			continue
		}

		break
	}

	if len(conditions) == 1 {
		return conditions[0]
	}
	return &BoolFilter{Operator: "AND", Conditions: conditions}
}

func (p *parser) parseNotExpr() Filter {
	if p.current.typ == tokenNOT {
		p.advance()
		operand := p.parseNotExpr()
		if operand == nil {
			return nil
		}
		return &BoolFilter{Operator: "NOT", Conditions: []Filter{operand}}
	}
	return p.parsePrimary()
}

func (p *parser) parsePrimary() Filter {
	if p.current.typ == tokenLParen {
		p.advance()
		expr := p.parseOrExpr()
		if p.current.typ == tokenRParen {
			p.advance()
		}
		return expr
	}

	return p.parseTerm()
}

func (p *parser) parseTerm() Filter {
	if p.current.typ != tokenWord {
		return nil
	}

	value := p.current.value
	p.advance()

	return parseFieldValue(value)
}

// parseFieldValue parses a single term like "from:alice" or "hello".
func parseFieldValue(term string) Filter {
	// Check for field:value syntax
	colonIdx := strings.Index(term, ":")
	if colonIdx > 0 {
		field := strings.ToLower(term[:colonIdx])
		value := term[colonIdx+1:]

		switch field {
		case "from":
			return &TextFilter{Field: "from", Value: value}
		case "to":
			return &TextFilter{Field: "to", Value: value}
		case "cc":
			return &TextFilter{Field: "cc", Value: value}
		case "bcc":
			return &TextFilter{Field: "bcc", Value: value}
		case "subject":
			return &TextFilter{Field: "subject", Value: value}
		case "body":
			return &TextFilter{Field: "body", Value: value}
		case "has":
			if strings.ToLower(value) == "attachment" {
				return &HasAttachmentFilter{Value: true}
			}
		case "is":
			switch strings.ToLower(value) {
			case "unread":
				return &TextFilter{Field: "notKeyword", Value: "$seen"}
			case "read":
				return &TextFilter{Field: "hasKeyword", Value: "$seen"}
			case "flagged", "starred":
				return &TextFilter{Field: "hasKeyword", Value: "$flagged"}
			case "unflagged", "unstarred":
				return &TextFilter{Field: "notKeyword", Value: "$flagged"}
			case "draft":
				return &TextFilter{Field: "hasKeyword", Value: "$draft"}
			case "answered":
				return &TextFilter{Field: "hasKeyword", Value: "$answered"}
			}
		case "in", "folder", "mailbox":
			return &TextFilter{Field: "inMailbox", Value: value}
		case "before":
			return &TextFilter{Field: "before", Value: value}
		case "after":
			return &TextFilter{Field: "after", Value: value}
		}
	}

	// Plain text search
	return &TextFilter{Field: "text", Value: term}
}

// ParseQuery parses a query string into a JMAP filter.
// Returns nil if the query is empty.
func ParseQuery(query string) Filter {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}
	p := newParser(query)
	return p.parse()
}

// containsBooleanOperators checks if a query contains boolean operators.
func containsBooleanOperators(query string) bool {
	tokens := newTokenizer(query)
	for {
		tok := tokens.nextToken()
		if tok.typ == tokenEOF {
			break
		}
		if tok.typ == tokenAND || tok.typ == tokenOR || tok.typ == tokenNOT || tok.typ == tokenLParen {
			return true
		}
	}
	return false
}
