package icu

import "strings"

// NodeType distinguishes different kinds of ICU message components.
type NodeType int

const (
	NodeText NodeType = iota
	NodeArgument
	NodePlural
	NodeSelect
	NodeSelectOrdinal
)

// Node represents an element in the ICU MessageFormat syntax tree.
type Node interface {
	Type() NodeType
	String() string
	Variables() []string
}

// TextNode represents raw literal text.
type TextNode struct {
	Text string
}

func (n *TextNode) Type() NodeType     { return NodeText }
func (n *TextNode) String() string     { return n.Text }
func (n *TextNode) Variables() []string { return nil }

// ArgumentNode represents a simple variable substitution like {name} or {count}.
type ArgumentNode struct {
	Name string
}

func (n *ArgumentNode) Type() NodeType     { return NodeArgument }
func (n *ArgumentNode) String() string     { return "{" + n.Name + "}" }
func (n *ArgumentNode) Variables() []string { return []string{n.Name} }

// CaseClause represents a branch in a plural or select structure.
type CaseClause struct {
	Selector string
	Message  *Message
}

// PluralNode represents plural selection: {count, plural, one{# item} other{# items}}.
type PluralNode struct {
	Name    string
	Offset  int
	Cases   []CaseClause
	Ordinal bool
}

func (n *PluralNode) Type() NodeType {
	if n.Ordinal {
		return NodeSelectOrdinal
	}
	return NodePlural
}

func (n *PluralNode) String() string {
	var sb strings.Builder
	sb.WriteString("{" + n.Name + ", ")
	if n.Ordinal {
		sb.WriteString("selectordinal, ")
	} else {
		sb.WriteString("plural, ")
	}
	for i, c := range n.Cases {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(c.Selector + "{" + c.Message.String() + "}")
	}
	sb.WriteString("}")
	return sb.String()
}

func (n *PluralNode) Variables() []string {
	vars := []string{n.Name}
	for _, c := range n.Cases {
		if c.Message != nil {
			vars = append(vars, c.Message.Variables()...)
		}
	}
	return deduplicate(vars)
}

// SelectNode represents gender/case selection: {gender, select, male{he} female{she} other{they}}.
type SelectNode struct {
	Name  string
	Cases []CaseClause
}

func (n *SelectNode) Type() NodeType { return NodeSelect }

func (n *SelectNode) String() string {
	var sb strings.Builder
	sb.WriteString("{" + n.Name + ", select, ")
	for i, c := range n.Cases {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(c.Selector + "{" + c.Message.String() + "}")
	}
	sb.WriteString("}")
	return sb.String()
}

func (n *SelectNode) Variables() []string {
	vars := []string{n.Name}
	for _, c := range n.Cases {
		if c.Message != nil {
			vars = append(vars, c.Message.Variables()...)
		}
	}
	return deduplicate(vars)
}

// Message is an ordered sequence of ICU AST nodes.
type Message struct {
	Nodes []Node
}

func (m *Message) String() string {
	var sb strings.Builder
	for _, n := range m.Nodes {
		sb.WriteString(n.String())
	}
	return sb.String()
}

func (m *Message) Variables() []string {
	var vars []string
	for _, n := range m.Nodes {
		vars = append(vars, n.Variables()...)
	}
	return deduplicate(vars)
}

func deduplicate(items []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, item := range items {
		if !seen[item] && item != "" {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}
