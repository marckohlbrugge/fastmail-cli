package jmap

import (
	"encoding/json"
	"testing"
)

func TestParseQuery_SimpleText(t *testing.T) {
	filter := ParseQuery("hello")
	if filter == nil {
		t.Fatal("expected non-nil filter")
	}

	jmap := filter.ToJMAP()
	if jmap["text"] != "hello" {
		t.Errorf("expected text=hello, got %v", jmap)
	}
}

func TestParseQuery_FieldFilter(t *testing.T) {
	tests := []struct {
		query string
		field string
		value string
	}{
		{"from:alice", "from", "alice"},
		{"to:bob", "to", "bob"},
		{"subject:meeting", "subject", "meeting"},
		{"cc:charlie", "cc", "charlie"},
		{"bcc:dave", "bcc", "dave"},
		{"body:important", "body", "important"},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			filter := ParseQuery(tt.query)
			if filter == nil {
				t.Fatal("expected non-nil filter")
			}

			jmap := filter.ToJMAP()
			if jmap[tt.field] != tt.value {
				t.Errorf("expected %s=%s, got %v", tt.field, tt.value, jmap)
			}
		})
	}
}

func TestParseQuery_SpecialFilters(t *testing.T) {
	tests := []struct {
		query string
		field string
		value interface{}
	}{
		{"has:attachment", "hasAttachment", true},
		{"is:unread", "notKeyword", "$seen"},
		{"is:read", "hasKeyword", "$seen"},
		{"is:flagged", "hasKeyword", "$flagged"},
		{"is:starred", "hasKeyword", "$flagged"},
		{"is:unflagged", "notKeyword", "$flagged"},
		{"is:draft", "hasKeyword", "$draft"},
		{"is:answered", "hasKeyword", "$answered"},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			filter := ParseQuery(tt.query)
			if filter == nil {
				t.Fatal("expected non-nil filter")
			}

			jmap := filter.ToJMAP()
			if jmap[tt.field] != tt.value {
				t.Errorf("expected %s=%v, got %v", tt.field, tt.value, jmap)
			}
		})
	}
}

func TestParseQuery_OROperator(t *testing.T) {
	filter := ParseQuery("hiring OR discount")
	if filter == nil {
		t.Fatal("expected non-nil filter")
	}

	jmap := filter.ToJMAP()
	if jmap["operator"] != "OR" {
		t.Errorf("expected operator=OR, got %v", jmap)
	}

	conditions, ok := jmap["conditions"].([]map[string]interface{})
	if !ok || len(conditions) != 2 {
		t.Errorf("expected 2 conditions, got %v", jmap["conditions"])
	}
	if conditions[0]["text"] != "hiring" {
		t.Errorf("expected first condition text=hiring, got %v", conditions[0])
	}
	if conditions[1]["text"] != "discount" {
		t.Errorf("expected second condition text=discount, got %v", conditions[1])
	}
}

func TestParseQuery_ANDOperator(t *testing.T) {
	filter := ParseQuery("from:alice AND subject:meeting")
	if filter == nil {
		t.Fatal("expected non-nil filter")
	}

	jmap := filter.ToJMAP()
	if jmap["operator"] != "AND" {
		t.Errorf("expected operator=AND, got %v", jmap)
	}

	conditions, ok := jmap["conditions"].([]map[string]interface{})
	if !ok || len(conditions) != 2 {
		t.Errorf("expected 2 conditions, got %v", jmap["conditions"])
	}
	if conditions[0]["from"] != "alice" {
		t.Errorf("expected first condition from=alice, got %v", conditions[0])
	}
	if conditions[1]["subject"] != "meeting" {
		t.Errorf("expected second condition subject=meeting, got %v", conditions[1])
	}
}

func TestParseQuery_ImplicitAND(t *testing.T) {
	filter := ParseQuery("from:alice to:bob")
	if filter == nil {
		t.Fatal("expected non-nil filter")
	}

	jmap := filter.ToJMAP()
	if jmap["operator"] != "AND" {
		t.Errorf("expected operator=AND, got %v", jmap)
	}

	conditions, ok := jmap["conditions"].([]map[string]interface{})
	if !ok || len(conditions) != 2 {
		t.Errorf("expected 2 conditions, got %v", jmap["conditions"])
	}
}

func TestParseQuery_NOTOperator(t *testing.T) {
	filter := ParseQuery("NOT spam")
	if filter == nil {
		t.Fatal("expected non-nil filter")
	}

	jmap := filter.ToJMAP()
	if jmap["operator"] != "NOT" {
		t.Errorf("expected operator=NOT, got %v", jmap)
	}

	conditions, ok := jmap["conditions"].([]map[string]interface{})
	if !ok || len(conditions) != 1 {
		t.Errorf("expected 1 condition, got %v", jmap["conditions"])
	}
	if conditions[0]["text"] != "spam" {
		t.Errorf("expected condition text=spam, got %v", conditions[0])
	}
}

func TestParseQuery_Parentheses(t *testing.T) {
	filter := ParseQuery("(a OR b) AND c")
	if filter == nil {
		t.Fatal("expected non-nil filter")
	}

	jmap := filter.ToJMAP()
	if jmap["operator"] != "AND" {
		t.Errorf("expected operator=AND, got %v", jmap)
	}

	conditions, ok := jmap["conditions"].([]map[string]interface{})
	if !ok || len(conditions) != 2 {
		t.Errorf("expected 2 conditions, got %v", jmap["conditions"])
	}

	// First condition should be OR
	if conditions[0]["operator"] != "OR" {
		t.Errorf("expected first condition operator=OR, got %v", conditions[0])
	}

	// Second condition should be text=c
	if conditions[1]["text"] != "c" {
		t.Errorf("expected second condition text=c, got %v", conditions[1])
	}
}

func TestParseQuery_MixedOperators(t *testing.T) {
	filter := ParseQuery("from:alice OR (subject:meeting AND NOT is:unread)")
	if filter == nil {
		t.Fatal("expected non-nil filter")
	}

	jmap := filter.ToJMAP()
	if jmap["operator"] != "OR" {
		t.Errorf("expected operator=OR, got %v", jmap)
	}

	conditions, ok := jmap["conditions"].([]map[string]interface{})
	if !ok || len(conditions) != 2 {
		t.Errorf("expected 2 conditions, got %v", jmap["conditions"])
	}

	// First condition: from:alice
	if conditions[0]["from"] != "alice" {
		t.Errorf("expected first condition from=alice, got %v", conditions[0])
	}

	// Second condition: AND filter
	if conditions[1]["operator"] != "AND" {
		t.Errorf("expected second condition operator=AND, got %v", conditions[1])
	}
}

func TestParseQuery_CaseInsensitiveOperators(t *testing.T) {
	tests := []struct {
		query    string
		operator string
	}{
		{"a OR b", "OR"},
		{"a or b", "OR"},
		{"a Or b", "OR"},
		{"a AND b", "AND"},
		{"a and b", "AND"},
		{"NOT a", "NOT"},
		{"not a", "NOT"},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			filter := ParseQuery(tt.query)
			if filter == nil {
				t.Fatal("expected non-nil filter")
			}

			jmap := filter.ToJMAP()
			if jmap["operator"] != tt.operator {
				t.Errorf("expected operator=%s, got %v", tt.operator, jmap)
			}
		})
	}
}

func TestParseQuery_QuotedStrings(t *testing.T) {
	tests := []struct {
		query string
		field string
		value string
	}{
		{`subject:"hello world"`, "subject", "hello world"},
		{`from:"Alice Smith"`, "from", "Alice Smith"},
		{`"exact phrase"`, "text", "exact phrase"},
		{`subject:'single quotes'`, "subject", "single quotes"},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			filter := ParseQuery(tt.query)
			if filter == nil {
				t.Fatal("expected non-nil filter")
			}

			jmap := filter.ToJMAP()
			if jmap[tt.field] != tt.value {
				t.Errorf("expected %s=%s, got %v", tt.field, tt.value, jmap)
			}
		})
	}
}

func TestParseQuery_Empty(t *testing.T) {
	filter := ParseQuery("")
	if filter != nil {
		t.Errorf("expected nil filter for empty query, got %v", filter)
	}

	filter = ParseQuery("   ")
	if filter != nil {
		t.Errorf("expected nil filter for whitespace query, got %v", filter)
	}
}

func TestParseQuery_MultipleORs(t *testing.T) {
	filter := ParseQuery("a OR b OR c")
	if filter == nil {
		t.Fatal("expected non-nil filter")
	}

	jmap := filter.ToJMAP()
	if jmap["operator"] != "OR" {
		t.Errorf("expected operator=OR, got %v", jmap)
	}

	conditions, ok := jmap["conditions"].([]map[string]interface{})
	if !ok || len(conditions) != 3 {
		t.Errorf("expected 3 conditions, got %v", jmap["conditions"])
	}
}

func TestParseQuery_OperatorPrecedence(t *testing.T) {
	// NOT has highest precedence, then AND, then OR
	// "a OR b AND NOT c" should be "a OR (b AND (NOT c))"
	filter := ParseQuery("a OR b AND NOT c")
	if filter == nil {
		t.Fatal("expected non-nil filter")
	}

	jmap := filter.ToJMAP()
	if jmap["operator"] != "OR" {
		t.Errorf("expected top-level operator=OR, got %v", jmap)
	}

	conditions, ok := jmap["conditions"].([]map[string]interface{})
	if !ok || len(conditions) != 2 {
		t.Errorf("expected 2 conditions, got %v", jmap["conditions"])
	}

	// First: text=a
	if conditions[0]["text"] != "a" {
		t.Errorf("expected first condition text=a, got %v", conditions[0])
	}

	// Second: AND of b and NOT c
	if conditions[1]["operator"] != "AND" {
		t.Errorf("expected second condition operator=AND, got %v", conditions[1])
	}
}

func TestContainsBooleanOperators(t *testing.T) {
	tests := []struct {
		query    string
		expected bool
	}{
		{"hello", false},
		{"from:alice", false},
		{"a OR b", true},
		{"a AND b", true},
		{"NOT a", true},
		{"(a)", true},
		{"from:alice to:bob", false}, // implicit AND doesn't count
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := containsBooleanOperators(tt.query)
			if result != tt.expected {
				t.Errorf("containsBooleanOperators(%q) = %v, want %v", tt.query, result, tt.expected)
			}
		})
	}
}

func TestParseQuery_ComplexNested(t *testing.T) {
	// Test a complex nested query
	filter := ParseQuery("(from:alice OR from:bob) AND (subject:meeting OR subject:call) AND NOT is:unread")
	if filter == nil {
		t.Fatal("expected non-nil filter")
	}

	// Verify it marshals to valid JSON
	jmap := filter.ToJMAP()
	_, err := json.Marshal(jmap)
	if err != nil {
		t.Errorf("failed to marshal filter to JSON: %v", err)
	}

	// Top level should be AND
	if jmap["operator"] != "AND" {
		t.Errorf("expected top-level operator=AND, got %v", jmap)
	}

	conditions, ok := jmap["conditions"].([]map[string]interface{})
	if !ok || len(conditions) != 3 {
		t.Errorf("expected 3 conditions, got %v", jmap["conditions"])
	}
}

func TestParseQuery_MailboxFilters(t *testing.T) {
	tests := []struct {
		query string
		value string
	}{
		{"in:inbox", "inbox"},
		{"folder:archive", "archive"},
		{"mailbox:drafts", "drafts"},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			filter := ParseQuery(tt.query)
			if filter == nil {
				t.Fatal("expected non-nil filter")
			}

			jmap := filter.ToJMAP()
			if jmap["inMailbox"] != tt.value {
				t.Errorf("expected inMailbox=%s, got %v", tt.value, jmap)
			}
		})
	}
}

func TestParseQuery_DateFilters(t *testing.T) {
	tests := []struct {
		query string
		field string
		value string
	}{
		{"before:2024-01-01", "before", "2024-01-01"},
		{"after:2024-01-01", "after", "2024-01-01"},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			filter := ParseQuery(tt.query)
			if filter == nil {
				t.Fatal("expected non-nil filter")
			}

			jmap := filter.ToJMAP()
			if jmap[tt.field] != tt.value {
				t.Errorf("expected %s=%s, got %v", tt.field, tt.value, jmap)
			}
		})
	}
}
