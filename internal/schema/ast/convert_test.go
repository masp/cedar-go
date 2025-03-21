package ast_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/schema/ast"
	"github.com/cedar-policy/cedar-go/internal/schema/parser"
)

func TestConvertHumanToJson(t *testing.T) {
	// Generate testdata/test_want.json by running:
	// 	cedar translate-schema --direction human-to-json -s testdata/test.cedarschema
	exampleHuman, err := os.ReadFile("testdata/convert/test.cedarschema")
	if err != nil {
		t.Fatalf("Error reading example schema: %v", err)
	}
	schema, err := parser.ParseFile("<test>", exampleHuman)
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

	want, err := os.ReadFile("testdata/convert/test_want.json")
	if err != nil {
		t.Fatalf("Error reading example JSON schema: %v", err)
	}
	ok, err := jsonEq(want, got.Bytes())
	if err != nil {
		t.Fatalf("Error comparing JSON: %v", err)
	}
	if !ok {
		if err := os.WriteFile("testdata/convert/test_got.json", got.Bytes(), 0644); err != nil {
			t.Logf("Error writing testdata/convert/test_got.json: %v", err)
		}
		t.Errorf("Schema does not match original, compare testdata/convert/test_want.json and testdata/convert/test_got.json")
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
