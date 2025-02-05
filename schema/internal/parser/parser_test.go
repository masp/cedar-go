package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/schema/ast"
)

func TestParseSimple(t *testing.T) {
	tests := []string{
		`namespace Demo {
}
`,
		`namespace Demo {
    entity User in UserGroup = {
        name: Demo::id,
        "department": UserGroup,
    };
}
`,
		`namespace PhotoFlash {
    entity User in UserGroup = {
        "department": String,
        "jobLevel": Long,
    };
}
`,
	}

	for _, test := range tests {
		schema, err := ParseFile("<test>", []byte(test))
		if err != nil {
			t.Fatalf("Error parsing simple schema: %v", err)
		}

		var buf bytes.Buffer
		ast.Format(schema, &buf)
		if strings.TrimSpace(buf.String()) != strings.TrimSpace(test) {
			t.Errorf("Parsed and formatted schema does not match:\nBefore\n%s\n=========================================\nAfter\n%s\n=========================================", test, buf.String())
		}
	}
}

func TestParseExample(t *testing.T) {
	example, err := os.ReadFile("testdata/format_test.cedarschema")
	if err != nil {
		t.Fatalf("open testfile: %v", err)
	}

	schema, err := ParseFile("<test>", []byte(example))
	if err != nil {
		t.Fatalf("Error parsing example schema: %v", err)
	}

	var buf bytes.Buffer
	var astbuf bytes.Buffer
	ast.Fprint(&astbuf, schema, nil)
	ast.Format(schema, &buf)
	if buf.String() != string(example) {
		t.Errorf("Parsed schema does not match original:\n%s\n=========================================\n%s\n=========================================", example, buf.String())
	}
}

func TestParserHasErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "missing closing bracket",
			input: `namespace PhotoFlash {`,
			want:  `<test>:1:23: expected }, got EOF`,
		},
		// {
		// 	name:  "missing entity name",
		// 	input: `namespace PhotoFlash { entity { "department": String }; }`,
		// 	want:  `<test>:1:31: expected identifier, got {`,
		// },
	}

	for _, test := range tests {
		_, err := ParseFile("<test>", []byte(test.input))
		if err == nil {
			t.Fatalf("Expected error parsing schema, got none")
		}
		if err.Error() != test.want {
			t.Errorf("Expected error %q, got %q", test.want, err.Error())
		}
	}
}

func TestConvertHumanToJson(t *testing.T) {
	// Generate testdata/convert_test.json by running:
	// 	cedar translate-schema --direction human-to-json -s testdata/convert_test.cedarschema
	exampleHuman, err := os.ReadFile("testdata/convert_test.cedarschema")
	if err != nil {
		t.Fatalf("Error reading example schema: %v", err)
	}
	schema, err := ParseFile("<test>", exampleHuman)
	if err != nil {
		t.Fatalf("Error parsing example schema: %v", err)
	}

	jsonSchema, err := ast.Convert(schema)
	if err != nil {
		t.Fatalf("Error marshalling schema to JSON: %v", err)
	}

	var got bytes.Buffer
	enc := json.NewEncoder(&got)
	enc.SetIndent("", "    ")
	err = enc.Encode(jsonSchema)
	if err != nil {
		t.Fatalf("Error dumping JSON: %v", err)
	}

	want, err := os.ReadFile("testdata/convert_test_want.json")
	if err != nil {
		t.Fatalf("Error reading example JSON schema: %v", err)
	}
	ok, err := jsonEq(want, got.Bytes())
	if err != nil {
		t.Fatalf("Error comparing JSON: %v", err)
	}
	if !ok {
		os.WriteFile("testdata/convert_test_got.json", got.Bytes(), 0644)
		t.Errorf("Schema does not match original, compare schema/testdata/convert_test_want.json and schema/testdata/convert_test_got.json")
	}
}

func jsonEq(a, b []byte) (bool, error) {
	var j, j2 interface{}
	if err := json.Unmarshal(a, &j); err != nil {
		return false, fmt.Errorf("left: %w", err)
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false, fmt.Errorf("right: %w", err)
	}
	return reflect.DeepEqual(j2, j), nil
}
