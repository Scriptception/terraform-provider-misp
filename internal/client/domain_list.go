// Package client provides types and helpers for interacting with the MISP REST API.
//
// DomainList exists to absorb MISP's four-way inconsistency in how
// restricted_to_domain is serialised depending on which endpoint produced the
// response and whether the field has content:
//
//   - POST /admin/organisations/add (empty)    → JSON null
//   - GET  /organisations/view      (empty)    → JSON array []
//   - GET  /organisations/view      (populated)→ JSON array ["example.com"]
//   - POST /admin/organisations/edit (empty)   → JSON string "[]"
//   - POST /admin/organisations/edit (populated)→ JSON string "[\"example.com\"]"
//
// The edit endpoint double-encodes: it JSON-stringifies the array and then
// wraps that string in a JSON string. This is a MISP bug that we absorb here
// so the rest of the provider always works with a plain []string.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// DomainList is []string that tolerates MISP's four different JSON encodings
// of restricted_to_domain: null, JSON array, string-containing-"[]", and
// string-containing-a-JSON-array. Always marshals as a plain JSON array.
type DomainList []string

// UnmarshalJSON decodes any of the four shapes MISP uses for this field.
func (d *DomainList) UnmarshalJSON(data []byte) error {
	s := bytes.TrimSpace(data)
	// 1. null → nil slice
	if bytes.Equal(s, []byte("null")) {
		*d = nil
		return nil
	}
	// 2. Proper JSON array
	if len(s) > 0 && s[0] == '[' {
		var arr []string
		if err := json.Unmarshal(s, &arr); err != nil {
			return fmt.Errorf("misp: DomainList array decode: %w", err)
		}
		*d = arr
		return nil
	}
	// 3. Quoted string — either "" or "[]" or stringified JSON array
	if len(s) > 0 && s[0] == '"' {
		var inner string
		if err := json.Unmarshal(s, &inner); err != nil {
			return fmt.Errorf("misp: DomainList string decode: %w", err)
		}
		if inner == "" || inner == "[]" {
			*d = []string{}
			return nil
		}
		var arr []string
		if err := json.Unmarshal([]byte(inner), &arr); err != nil {
			return fmt.Errorf("misp: DomainList inner decode: %w (inner: %q)", err, inner)
		}
		*d = arr
		return nil
	}
	return fmt.Errorf("misp: cannot decode %s as DomainList", s)
}

// MarshalJSON always emits a plain JSON array (MISP accepts this on input).
func (d DomainList) MarshalJSON() ([]byte, error) {
	if len(d) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal([]string(d))
}
