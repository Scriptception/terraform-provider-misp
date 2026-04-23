package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) (*Client, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c, err := New(Config{URL: srv.URL, APIKey: "test-key"})
	if err != nil {
		srv.Close()
		t.Fatalf("New: %v", err)
	}
	return c, srv.Close
}

func TestNew_validation(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
		want string
	}{
		{"missing url", Config{APIKey: "k"}, "url is required"},
		{"missing key", Config{URL: "https://x"}, "api_key is required"},
		{"bad url", Config{URL: "not-a-url", APIKey: "k"}, "must include scheme"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.cfg)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestDo_sendsAuthAndAcceptHeaders(t *testing.T) {
	c, stop := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "test-key" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q", got)
		}
		_, _ = w.Write([]byte(`{}`))
	})
	defer stop()

	if err := c.do(context.Background(), "GET", "/ping", nil, &struct{}{}); err != nil {
		t.Fatalf("do: %v", err)
	}
}

func TestDo_nonSuccessReturnsAPIError(t *testing.T) {
	c, stop := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"nope"}`))
	})
	defer stop()

	err := c.do(context.Background(), "GET", "/missing", nil, nil)
	if !IsNotFound(err) {
		t.Fatalf("want NotFound, got %v", err)
	}
}

func TestGetOrganisation_decodesEnvelope(t *testing.T) {
	c, stop := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"Organisation": map[string]any{"id": "7", "name": "ACME", "local": true},
		})
	})
	defer stop()

	got, err := c.GetOrganisation(context.Background(), "7")
	if err != nil {
		t.Fatalf("GetOrganisation: %v", err)
	}
	if got.ID != "7" || got.Name != "ACME" || !got.Local {
		t.Errorf("unexpected org: %+v", got)
	}
}

// GetTag must decode MISP's bare (non-enveloped) /tags/view response.
// Regression: a prior version decoded into a tagEnvelope and silently
// produced an all-zero Tag, blanking state.ID on refresh.
func TestGetTag_decodesFlatResponse(t *testing.T) {
	c, stop := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tags/view/7" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id":"7","name":"tlp:amber","colour":"#FFC000","exportable":true,"hide_tag":false,"org_id":"0","user_id":"0","local_only":false}`))
	})
	defer stop()

	got, err := c.GetTag(context.Background(), "7")
	if err != nil {
		t.Fatalf("GetTag: %v", err)
	}
	if got.ID != "7" || got.Name != "tlp:amber" || got.Colour != "#FFC000" || !got.Exportable {
		t.Errorf("unexpected tag: %+v", got)
	}
}

// CreateTag must decode the enveloped /tags/add response.
func TestCreateTag_decodesEnvelope(t *testing.T) {
	c, stop := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"Tag": map[string]any{"id": "9", "name": "x", "exportable": true},
		})
	})
	defer stop()

	got, err := c.CreateTag(context.Background(), Tag{Name: "x"})
	if err != nil {
		t.Fatalf("CreateTag: %v", err)
	}
	if got.ID != "9" || got.Name != "x" {
		t.Errorf("unexpected tag: %+v", got)
	}
}

// TestFlexString_unmarshal covers all three JSON shapes FlexString must accept:
// quoted strings, bare numbers, and JSON booleans.
func TestFlexString_unmarshal(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"quoted string", `"hello"`, "hello"},
		{"bare integer", `42`, "42"},
		{"bare float", `3.14`, "3.14"},
		{"bool true", `true`, "true"},
		{"bool false", `false`, "false"},
		{"null", `null`, ""},
		{"empty string", `""`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var s FlexString
			if err := json.Unmarshal([]byte(tc.input), &s); err != nil {
				t.Fatalf("UnmarshalJSON(%s): %v", tc.input, err)
			}
			if got := s.String(); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// TestGetSetting_decodesFlatResponse verifies that GetSetting correctly
// decodes the flat (non-enveloped) JSON returned by MISP for both string and
// boolean value fields.
func TestGetSetting_decodesFlatResponse(t *testing.T) {
	cases := []struct {
		name      string
		body      string
		wantName  string
		wantValue string
		wantType  string
	}{
		{
			name:      "string value",
			body:      `{"name":"MISP.baseurl","value":"https://misp.example.com","type":"string","description":"The base url","level":0}`,
			wantName:  "MISP.baseurl",
			wantValue: "https://misp.example.com",
			wantType:  "string",
		},
		{
			name:      "boolean value true",
			body:      `{"name":"MISP.disable_emailing","value":true,"type":"boolean","description":"Disable emailing","level":0}`,
			wantName:  "MISP.disable_emailing",
			wantValue: "true",
			wantType:  "boolean",
		},
		{
			name:      "boolean value false",
			body:      `{"name":"MISP.disable_emailing","value":false,"type":"boolean","description":"Disable emailing","level":0}`,
			wantName:  "MISP.disable_emailing",
			wantValue: "false",
			wantType:  "boolean",
		},
		{
			name:      "numeric value",
			body:      `{"name":"MISP.correlation_engine","value":5000,"type":"numeric","description":"Threshold","level":1}`,
			wantName:  "MISP.correlation_engine",
			wantValue: "5000",
			wantType:  "numeric",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, stop := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(tc.body))
			})
			defer stop()

			got, err := c.GetSetting(context.Background(), tc.wantName)
			if err != nil {
				t.Fatalf("GetSetting: %v", err)
			}
			if got.Name != tc.wantName {
				t.Errorf("Name: got %q, want %q", got.Name, tc.wantName)
			}
			if got.Value.String() != tc.wantValue {
				t.Errorf("Value: got %q, want %q", got.Value.String(), tc.wantValue)
			}
			if got.Type != tc.wantType {
				t.Errorf("Type: got %q, want %q", got.Type, tc.wantType)
			}
		})
	}
}

// TestUpdateSetting_sendsStringBody verifies that UpdateSetting always sends
// the value as a JSON string (even for values that look like booleans or
// numbers), matching MISP's API expectation.
func TestUpdateSetting_sendsStringBody(t *testing.T) {
	values := []string{"true", "false", "5000", "https://example.com"}
	for _, v := range values {
		t.Run(v, func(t *testing.T) {
			c, stop := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					Value string `json:"value"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode request body: %v", err)
				}
				if body.Value != v {
					t.Errorf("body.value = %q, want %q", body.Value, v)
				}
				_, _ = w.Write([]byte(`{"saved":true,"success":true}`))
			})
			defer stop()

			if err := c.UpdateSetting(context.Background(), "MISP.test", v); err != nil {
				t.Fatalf("UpdateSetting: %v", err)
			}
		})
	}
}
