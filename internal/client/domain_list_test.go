package client

import (
	"encoding/json"
	"testing"
)

// helper: unmarshal raw JSON bytes into a DomainList.
func unmarshalDomainList(t *testing.T, raw string) (DomainList, error) {
	t.Helper()
	var d DomainList
	err := json.Unmarshal([]byte(raw), &d)
	return d, err
}

func TestDomainList_unmarshal_null(t *testing.T) {
	d, err := unmarshalDomainList(t, `null`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != nil {
		t.Fatalf("expected nil slice, got %v", d)
	}
}

func TestDomainList_unmarshal_emptyArray(t *testing.T) {
	d, err := unmarshalDomainList(t, `[]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(d) != 0 {
		t.Fatalf("expected empty slice, got %v", d)
	}
}

func TestDomainList_unmarshal_populatedArray(t *testing.T) {
	d, err := unmarshalDomainList(t, `["a","b"]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(d) != 2 || d[0] != "a" || d[1] != "b" {
		t.Fatalf("expected [a b], got %v", d)
	}
}

func TestDomainList_unmarshal_quotedEmpty(t *testing.T) {
	// MISP edit endpoint returns "[]" as the empty state.
	d, err := unmarshalDomainList(t, `"[]"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(d) != 0 {
		t.Fatalf("expected empty slice, got %v", d)
	}
}

func TestDomainList_unmarshal_quotedPopulated(t *testing.T) {
	// MISP edit endpoint double-encodes: the JSON value is a string whose
	// content is itself a JSON array literal.
	d, err := unmarshalDomainList(t, `"[\"a\",\"b\"]"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(d) != 2 || d[0] != "a" || d[1] != "b" {
		t.Fatalf("expected [a b], got %v", d)
	}
}

func TestDomainList_unmarshal_quotedBlank(t *testing.T) {
	// Empty quoted string ("") should also yield an empty (non-nil) slice.
	d, err := unmarshalDomainList(t, `""`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(d) != 0 {
		t.Fatalf("expected empty slice, got %v", d)
	}
}

func TestDomainList_unmarshal_invalid(t *testing.T) {
	_, err := unmarshalDomainList(t, `{"bad":"input"}`)
	if err == nil {
		t.Fatal("expected error for object input, got nil")
	}
}

// --- MarshalJSON tests ---

func TestDomainList_marshal_nil(t *testing.T) {
	var d DomainList // nil
	out, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "[]" {
		t.Fatalf("expected [], got %s", out)
	}
}

func TestDomainList_marshal_empty(t *testing.T) {
	d := DomainList{}
	out, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "[]" {
		t.Fatalf("expected [], got %s", out)
	}
}

func TestDomainList_marshal_populated(t *testing.T) {
	d := DomainList{"a", "b"}
	out, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != `["a","b"]` {
		t.Fatalf(`expected ["a","b"], got %s`, out)
	}
}
