package icu_test

import (
	"testing"

	"github.com/alexandrmotologa/localecraft/internal/icu"
)

func TestParseSimpleMessage(t *testing.T) {
	msg, err := icu.Parse("Hello, {name}! You have {unreadCount} unread messages.")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	vars := msg.Variables()
	if len(vars) != 2 || vars[0] != "name" || vars[1] != "unreadCount" {
		t.Fatalf("unexpected variables: %v", vars)
	}
}

func TestParsePluralMessage(t *testing.T) {
	input := "{count, plural, =0{No files remaining} one{# file remaining} other{# files remaining}}"
	msg, err := icu.Parse(input)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if len(msg.Nodes) != 1 {
		t.Fatalf("expected 1 plural node, got %d", len(msg.Nodes))
	}

	pluralNode, ok := msg.Nodes[0].(*icu.PluralNode)
	if !ok {
		t.Fatalf("expected *icu.PluralNode, got %T", msg.Nodes[0])
	}

	if pluralNode.Name != "count" || len(pluralNode.Cases) != 3 {
		t.Fatalf("unexpected plural node state: %v", pluralNode)
	}
}

func TestValidationParity(t *testing.T) {
	source := "Welcome {userName}, you have {taskCount, plural, one{# task} other{# tasks}}."

	// 1. Valid target
	validTarget := "Willkommen {userName}, Sie haben {taskCount, plural, one{# Aufgabe} other{# Aufgaben}}."
	res := icu.Validate(source, validTarget, "de")
	if !res.Valid {
		t.Fatalf("expected valid target, got errors: %v", res.Errors)
	}

	// 2. Missing variable {taskCount}
	missingTarget := "Willkommen {userName}, Sie haben neue Aufgaben."
	resMissing := icu.Validate(source, missingTarget, "de")
	if resMissing.Valid {
		t.Fatalf("expected invalid result due to missing variable")
	}
	if len(resMissing.MissingVars) != 1 || resMissing.MissingVars[0] != "taskCount" {
		t.Fatalf("expected taskCount to be flagged missing, got: %v", resMissing.MissingVars)
	}

	// 3. Hallucinated / unexpected variable {userEmail}
	extraTarget := "Willkommen {userName} ({userEmail}), Sie haben {taskCount, plural, one{# Aufgabe} other{# Aufgaben}}."
	resExtra := icu.Validate(source, extraTarget, "de")
	if resExtra.Valid {
		t.Fatalf("expected invalid result due to extra variable")
	}
	if len(resExtra.UnexpectedVars) != 1 || resExtra.UnexpectedVars[0] != "userEmail" {
		t.Fatalf("expected userEmail to be flagged unexpected, got: %v", resExtra.UnexpectedVars)
	}

	// 4. Missing mandatory 'other' plural case
	brokenPluralTarget := "Willkommen {userName}, Sie haben {taskCount, plural, one{# Aufgabe}}."
	resBrokenPlural := icu.Validate(source, brokenPluralTarget, "de")
	if resBrokenPlural.Valid {
		t.Fatalf("expected invalid result due to missing 'other' case")
	}
}

func TestEvaluate(t *testing.T) {
	input := "{count, plural, =0{None} one{# item} other{# items}}"
	msg, err := icu.Parse(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// count = 0 -> "None"
	out0, err := icu.Evaluate(msg, map[string]any{"count": 0}, "en")
	if err != nil || out0 != "None" {
		t.Errorf("expected 'None', got '%s'", out0)
	}

	// count = 1 -> "1 item"
	out1, err := icu.Evaluate(msg, map[string]any{"count": 1}, "en")
	if err != nil || out1 != "1 item" {
		t.Errorf("expected '1 item', got '%s'", out1)
	}

	// count = 5 -> "5 items"
	out5, err := icu.Evaluate(msg, map[string]any{"count": 5}, "en")
	if err != nil || out5 != "5 items" {
		t.Errorf("expected '5 items', got '%s'", out5)
	}
}

func TestUnmatchedBracesError(t *testing.T) {
	badInput := "Hello, {name"
	_, err := icu.Parse(badInput)
	if err == nil {
		t.Fatalf("expected error for unclosed brace, got nil")
	}
}
