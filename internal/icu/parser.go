package icu

import (
	"fmt"
	"strings"
	"unicode"
)

// Parser parses ICU MessageFormat strings into an AST.
type Parser struct {
	input []rune
	pos   int
	len   int
}

// Parse parses a full ICU MessageFormat string.
func Parse(input string) (*Message, error) {
	p := &Parser{
		input: []rune(input),
		pos:   0,
		len:   len([]rune(input)),
	}
	return p.parseMessage(false)
}

func (p *Parser) parseMessage(inSubMessage bool) (*Message, error) {
	msg := &Message{Nodes: make([]Node, 0)}
	var textBuf strings.Builder

	flushText := func() {
		if textBuf.Len() > 0 {
			msg.Nodes = append(msg.Nodes, &TextNode{Text: textBuf.String()})
			textBuf.Reset()
		}
	}

	for p.pos < p.len {
		ch := p.input[p.pos]

		if inSubMessage && ch == '}' {
			// End of case sub-message
			flushText()
			return msg, nil
		}

		if ch == '\'' {
			// Escaping logic
			if p.pos+1 < p.len && p.input[p.pos+1] == '\'' {
				textBuf.WriteRune('\'')
				p.pos += 2
				continue
			}
			if p.pos+1 < p.len && (p.input[p.pos+1] == '{' || p.input[p.pos+1] == '}') {
				// Quoted literal brace
				p.pos++
				textBuf.WriteRune(p.input[p.pos])
				p.pos++
				if p.pos < p.len && p.input[p.pos] == '\'' {
					p.pos++ // consume closing quote if present
				}
				continue
			}
			textBuf.WriteRune('\'')
			p.pos++
			continue
		}

		if ch == '{' {
			flushText()
			node, err := p.parseTag()
			if err != nil {
				return nil, err
			}
			msg.Nodes = append(msg.Nodes, node)
			continue
		}

		if ch == '}' && !inSubMessage {
			return nil, fmt.Errorf("unexpected closing brace at position %d", p.pos)
		}

		textBuf.WriteRune(ch)
		p.pos++
	}

	flushText()
	return msg, nil
}

func (p *Parser) parseTag() (Node, error) {
	startPos := p.pos
	p.pos++ // skip '{'
	p.skipWhitespace()

	// Read identifier / argument name
	argName := p.readIdentifier()
	if argName == "" {
		return nil, fmt.Errorf("empty argument name inside brace at position %d", startPos)
	}

	p.skipWhitespace()
	if p.pos >= p.len {
		return nil, fmt.Errorf("unclosed brace for argument '{%s' starting at %d", argName, startPos)
	}

	if p.input[p.pos] == '}' {
		p.pos++ // consume '}'
		return &ArgumentNode{Name: argName}, nil
	}

	if p.input[p.pos] != ',' {
		return nil, fmt.Errorf("expected ',' or '}' after argument '%s' at position %d, got '%c'", argName, p.pos, p.input[p.pos])
	}
	p.pos++ // consume ','
	p.skipWhitespace()

	// Read tag type: plural, select, selectordinal, number, date, time
	tagType := p.readIdentifier()
	tagType = strings.ToLower(tagType)
	p.skipWhitespace()

	switch tagType {
	case "plural", "selectordinal":
		isOrdinal := tagType == "selectordinal"
		if p.pos >= p.len || p.input[p.pos] != ',' {
			return nil, fmt.Errorf("expected ',' after %s keyword at position %d", tagType, p.pos)
		}
		p.pos++ // consume ','
		p.skipWhitespace()

		cases, offset, err := p.parseCases()
		if err != nil {
			return nil, err
		}

		if p.pos >= p.len || p.input[p.pos] != '}' {
			return nil, fmt.Errorf("expected '}' closing %s for '%s' at position %d", tagType, argName, p.pos)
		}
		p.pos++ // consume '}'

		return &PluralNode{
			Name:    argName,
			Offset:  offset,
			Cases:   cases,
			Ordinal: isOrdinal,
		}, nil

	case "select":
		if p.pos >= p.len || p.input[p.pos] != ',' {
			return nil, fmt.Errorf("expected ',' after select keyword at position %d", p.pos)
		}
		p.pos++ // consume ','
		p.skipWhitespace()

		cases, _, err := p.parseCases()
		if err != nil {
			return nil, err
		}

		if p.pos >= p.len || p.input[p.pos] != '}' {
			return nil, fmt.Errorf("expected '}' closing select for '%s' at position %d", argName, p.pos)
		}
		p.pos++ // consume '}'

		return &SelectNode{
			Name:  argName,
			Cases: cases,
		}, nil

	default:
		// Simple format with style: e.g. {date, date, short} or {num, number, percent}
		// Treat as extended ArgumentNode representation
		var rest strings.Builder
		rest.WriteString(argName)
		rest.WriteString(", ")
		rest.WriteString(tagType)
		for p.pos < p.len && p.input[p.pos] != '}' {
			rest.WriteRune(p.input[p.pos])
			p.pos++
		}
		if p.pos < p.len && p.input[p.pos] == '}' {
			p.pos++
		}
		return &ArgumentNode{Name: rest.String()}, nil
	}
}

func (p *Parser) parseCases() ([]CaseClause, int, error) {
	var cases []CaseClause
	offset := 0

	for {
		p.skipWhitespace()
		if p.pos >= p.len || p.input[p.pos] == '}' {
			break
		}

		selector := p.readSelector()
		if selector == "" {
			return nil, 0, fmt.Errorf("expected selector or closing brace at position %d", p.pos)
		}

		if strings.HasPrefix(selector, "offset:") {
			// Offset specified
			var offsetVal int
			fmt.Sscanf(selector, "offset:%d", &offsetVal)
			offset = offsetVal
			continue
		}

		p.skipWhitespace()
		if p.pos >= p.len || p.input[p.pos] != '{' {
			return nil, 0, fmt.Errorf("expected '{' after selector '%s' at position %d", selector, p.pos)
		}
		p.pos++ // consume '{'

		subMsg, err := p.parseMessage(true)
		if err != nil {
			return nil, 0, err
		}

		if p.pos >= p.len || p.input[p.pos] != '}' {
			return nil, 0, fmt.Errorf("expected '}' ending case '%s' at position %d", selector, p.pos)
		}
		p.pos++ // consume '}'

		cases = append(cases, CaseClause{
			Selector: selector,
			Message:  subMsg,
		})
	}

	return cases, offset, nil
}

func (p *Parser) readIdentifier() string {
	start := p.pos
	for p.pos < p.len {
		ch := p.input[p.pos]
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '-' || ch == '.' {
			p.pos++
		} else {
			break
		}
	}
	return string(p.input[start:p.pos])
}

func (p *Parser) readSelector() string {
	start := p.pos
	for p.pos < p.len {
		ch := p.input[p.pos]
		if ch == '{' || unicode.IsSpace(ch) || ch == '}' {
			break
		}
		p.pos++
	}
	return string(p.input[start:p.pos])
}

func (p *Parser) skipWhitespace() {
	for p.pos < p.len && unicode.IsSpace(p.input[p.pos]) {
		p.pos++
	}
}
