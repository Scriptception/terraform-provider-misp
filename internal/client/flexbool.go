package client

import (
	"bytes"
	"fmt"
	"strconv"
)

// FlexBool is a bool that tolerates MISP's inconsistent JSON encoding. Different
// MISP endpoints (and often different MISP versions) return the same boolean
// field as `true/false`, `1/0`, or `"1"/"0"` — we've observed all three.
// FlexBool unmarshals any of them.
type FlexBool bool

// Bool returns the underlying value.
func (b FlexBool) Bool() bool { return bool(b) }

// UnmarshalJSON decodes JSON bool, number 0/1, or quoted string "0"/"1"/"true"/"false".
func (b *FlexBool) UnmarshalJSON(data []byte) error {
	s := string(bytes.TrimSpace(data))
	switch s {
	case "true":
		*b = true
	case "false", "null":
		*b = false
	case "0":
		*b = false
	case "1":
		*b = true
	default:
		// Try quoted forms.
		if unq, err := strconv.Unquote(s); err == nil {
			switch unq {
			case "true", "1":
				*b = true
				return nil
			case "false", "0", "":
				*b = false
				return nil
			}
		}
		return fmt.Errorf("misp: cannot decode %s as bool", s)
	}
	return nil
}

// MarshalJSON always emits a plain JSON bool (MISP accepts this shape on input).
func (b FlexBool) MarshalJSON() ([]byte, error) {
	if b {
		return []byte("true"), nil
	}
	return []byte("false"), nil
}

// FlexString is a string that tolerates JSON numbers. MISP frequently returns
// numeric-looking fields (rate_limit_count, nids_sid, ...) as either strings
// ("100") or bare numbers (100), depending on the endpoint. FlexString absorbs
// both.
type FlexString string

// String returns the underlying string.
func (s FlexString) String() string { return string(s) }

// UnmarshalJSON accepts JSON strings, JSON numbers, and JSON booleans.
// Booleans are normalised to "true" or "false" strings, which matches
// what MISP expects when the value is round-tripped back via the API.
func (s *FlexString) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		*s = ""
		return nil
	}
	if trimmed[0] == '"' {
		unq, err := strconv.Unquote(string(trimmed))
		if err != nil {
			return err
		}
		*s = FlexString(unq)
		return nil
	}
	// JSON boolean literals — normalise to "true"/"false" strings.
	if string(trimmed) == "true" {
		*s = "true"
		return nil
	}
	if string(trimmed) == "false" {
		*s = "false"
		return nil
	}
	// Bare JSON number — keep the raw digits as a string.
	*s = FlexString(trimmed)
	return nil
}

// MarshalJSON emits the value as a JSON string.
func (s FlexString) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(string(s))), nil
}
