package cedar

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestNewSchema(t *testing.T) {
	s := NewSchema()
	if s == nil {
		t.Error("NewSchema() returned nil")
	}
}

func TestSchemaCedarMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid schema",
			input: `namespace foo {
				action Bar appliesTo {
					principal: String,
					resource: String
				};
			}`,
			wantErr: false,
		},
		{
			name:    "empty schema",
			input:   "",
			wantErr: false,
		},
		{
			name: "invalid schema",
			input: `namespace foo {
				action Bar = {
					invalid syntax here
				};
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSchema()
			err := s.UnmarshalCedar("test.cedar", []byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalCedar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Test marshaling
			out, err := s.MarshalCedar()
			if err != nil {
				t.Errorf("MarshalCedar() error = %v", err)
				return
			}

			// For valid schemas, unmarshaling and marshaling should preserve content
			if !tt.wantErr {
				s2 := NewSchema()
				err = s2.UnmarshalCedar("test.cedar", out)
				if err != nil {
					t.Errorf("UnmarshalCedar() second pass error = %v", err)
					return
				}

				out2, err := s2.MarshalCedar()
				if err != nil {
					t.Errorf("MarshalCedar() second pass error = %v", err)
					return
				}

				if !reflect.DeepEqual(out, out2) {
					t.Errorf("Marshal/Unmarshal cycle produced different results:\nFirst: %s\nSecond: %s", out, out2)
				}
			}
		})
	}
}

func TestSchemaCedarMarshalEmpty(t *testing.T) {
	s := NewSchema()
	out, err := s.MarshalCedar()
	if err != nil {
		t.Errorf("MarshalCedar() error = %v", err)
		return
	}
	if len(out) != 0 {
		t.Errorf("MarshalCedar() produced non-empty output for empty schema")
	}
}

func TestSchemaJSONMarshalEmpty(t *testing.T) {
	s := NewSchema()
	out, err := s.MarshalJSON()
	if err != nil {
		t.Errorf("MarshalJSON() error = %v", err)
		return
	}
	if len(out) != 0 {
		t.Errorf("MarshalJSON() produced non-empty output for empty schema")
	}
}

func TestSchemaJSONMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid JSON schema",
			input: `{
				"entityTypes": {
					"User": {
						"shape": {
							"type": "Record",
							"attributes": {
								"name": {"type": "String"}
							}
						}
					}
				}
			}`,
			wantErr: false,
		},
		{
			name:    "empty JSON",
			input:   "{}",
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   "{invalid json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSchema()
			err := s.UnmarshalJSON([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Test marshaling
			out, err := s.MarshalJSON()
			if err != nil {
				t.Errorf("MarshalJSON() error = %v", err)
				return
			}

			// Verify JSON validity
			var raw interface{}
			if err := json.Unmarshal(out, &raw); err != nil {
				t.Errorf("MarshalJSON() produced invalid JSON: %v", err)
			}
		})
	}
}

func TestSchemaCrossFormatMarshaling(t *testing.T) {
	t.Run("JSON to Cedar marshaling not allowed", func(t *testing.T) {
		s := NewSchema()
		err := s.UnmarshalJSON([]byte(`{}`))
		if err != nil {
			t.Fatalf("UnmarshalJSON() error = %v", err)
		}

		_, err = s.MarshalCedar()
		if err == nil {
			t.Error("MarshalCedar() should return error after UnmarshalJSON")
		}
	})

	t.Run("Cedar to JSON marshaling allowed", func(t *testing.T) {
		s := NewSchema()
		err := s.UnmarshalCedar("test.cedar", []byte(`namespace test {}`))
		if err != nil {
			t.Fatalf("UnmarshalCedar() error = %v", err)
		}

		_, err = s.MarshalJSON()
		if err != nil {
			t.Errorf("MarshalJSON() error = %v", err)
		}
	})
}
