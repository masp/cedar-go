package token

import (
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

type Error struct {
	Pos Position
	Err error
}

func (e Error) Error() string {
	filename := e.Pos.Filename
	if filename == "" {
		filename = "<input>"
	}
	return fmt.Sprintf("%s:%d:%d: %v", filename, e.Pos.Line, e.Pos.Column, e.Err)
}

type ErrList []error

func (errs ErrList) Error() string {
	return errors.Join(errs...).Error()
}

func (errs ErrList) Sort() {
	errs = slices.DeleteFunc(errs, func(e1 error) bool { return e1 == nil })
	slices.SortFunc(errs, func(e1, e2 error) int {
		te1 := e1.(Error)
		te2 := e2.(Error)
		return te1.Pos.Offset - te2.Pos.Offset
	})
}

type Position = ast.Position

type Type int

const (
	INVALID Type = iota

	// Keywords
	NAMESPACE
	ENTITY
	ACTION
	TYPE
	IN
	TAGS
	APPLIES_TO
	PRINCIPAL
	RESOURCE
	CONTEXT

	// Reserved type names
	BOOL_TYPE
	LONG_TYPE
	STRING_TYPE
	SET_TYPE

	// Punctuation
	LEFTBRACE    // {
	RIGHTBRACE   // }
	LEFTBRACKET  // [
	RIGHTBRACKET // ]
	LEFTANGLE    // <
	RIGHTANGLE   // >
	COLON        // :
	SEMICOLON    // ;
	COMMA        // ,
	EQUALS       // =
	QUESTION     // ?
	DOUBLECOLON  // ::

	// Identifiers and literals
	IDENT  // Regular identifier
	STRING // String literal, quoted with ""

	// Comments
	COMMENT // // style comment

	// Special
	EOF // End of file
)

var types = [...]string{
	INVALID:    "INVALID",
	NAMESPACE:  "NAMESPACE",
	ENTITY:     "ENTITY",
	ACTION:     "ACTION",
	TYPE:       "TYPE",
	IN:         "IN",
	TAGS:       "TAGS",
	APPLIES_TO: "APPLIES_TO",
	PRINCIPAL:  "PRINCIPAL",
	RESOURCE:   "RESOURCE",
	CONTEXT:    "CONTEXT",

	BOOL_TYPE:   "BOOL_TYPE",
	LONG_TYPE:   "LONG_TYPE",
	STRING_TYPE: "STRING_TYPE",
	SET_TYPE:    "SET_TYPE",

	LEFTBRACE:    "LEFTBRACE",
	RIGHTBRACE:   "RIGHTBRACE",
	LEFTBRACKET:  "LEFTBRACKET",
	RIGHTBRACKET: "RIGHTBRACKET",
	LEFTANGLE:    "LEFTANGLE",
	RIGHTANGLE:   "RIGHTANGLE",
	COLON:        "COLON",
	SEMICOLON:    "SEMICOLON",
	COMMA:        "COMMA",
	EQUALS:       "EQUALS",
	QUESTION:     "QUESTION",
	DOUBLECOLON:  "DOUBLECOLON",

	IDENT:  "IDENT",
	STRING: "STRING",

	COMMENT: "COMMENT",

	EOF: "EOF",
}

func (tok Type) String() string {
	s := ""
	if 0 <= tok && tok < Type(len(types)) {
		s = types[tok]
	}
	if s == "" {
		s = "Token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}
