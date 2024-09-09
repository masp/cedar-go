package eval

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestPartial(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   *ast.Policy
		env  *Env
		out  *ast.Policy
		keep bool
	}{
		{"smokeTest",
			ast.Permit(),
			&Env{},
			ast.Permit(),
			true,
		},
		{"principalEqual",
			ast.Permit().PrincipalEq(types.NewEntityUID("Account", "42")),
			&Env{
				Principal: types.NewEntityUID("Account", "42"),
			},
			ast.Permit(),
			true,
		},
		{"principalNotEqual",
			ast.Permit().PrincipalEq(types.NewEntityUID("Account", "42")),
			&Env{
				Principal: types.NewEntityUID("Account", "Other"),
			},
			nil,
			false,
		},
		{"conditionOmitTrue",
			ast.Permit().When(ast.True()),
			&Env{},
			ast.Permit(),
			true,
		},
		{"conditionDropFalse",
			ast.Permit().When(ast.False()),
			&Env{},
			nil,
			false,
		},
		{"conditionDropError",
			ast.Permit().When(ast.Long(42).GreaterThan(ast.String("bananas"))),
			&Env{},
			nil,
			false,
		},
		{"conditionDropTypeError",
			ast.Permit().When(ast.Long(42)),
			&Env{},
			nil,
			false,
		},
		{"conditionKeepUnfolded",
			ast.Permit().When(ast.Context().GreaterThan(ast.Long(42))),
			&Env{Context: Variable("context")},
			ast.Permit().When(ast.Context().GreaterThan(ast.Long(42))),
			true,
		},
		{"conditionOmitTrueFolded",
			ast.Permit().When(ast.Context().GreaterThan(ast.Long(42))),
			&Env{
				Context: types.Long(43),
			},
			ast.Permit(),
			true,
		},
		{"conditionDropFalseFolded",
			ast.Permit().When(ast.Context().GreaterThan(ast.Long(42))),
			&Env{
				Context: types.Long(41),
			},
			nil,
			false,
		},
		{"conditionDropErrorFolded",
			ast.Permit().When(ast.Context().GreaterThan(ast.Long(42))),
			&Env{
				Context: types.String("bananas"),
			},
			nil,
			false,
		},
		{"contextVariableAccess",
			ast.Permit().When(ast.Context().Access("key").Equal(ast.Long(42))),
			&Env{
				Context: types.Record{
					"key": Variable("var"),
				},
			},
			ast.Permit().When(ast.Context().Access("key").Equal(ast.Long(42))),
			true,
		},

		{"ignorePermitContext",
			ast.Permit().When(ast.Context().Equal(ast.Long(42))),
			&Env{
				Context: Ignore(),
			},
			ast.Permit(),
			true,
		},
		{"ignoreForbidContext",
			ast.Forbid().When(ast.Context().Equal(ast.Long(42))),
			&Env{
				Context: Ignore(),
			},
			nil,
			false,
		},
		{"ignorePermitScope",
			ast.Permit().PrincipalEq(types.NewEntityUID("T", "42")),
			&Env{
				Principal: Ignore(),
			},
			ast.Permit(),
			true,
		},
		{"ignoreForbidScope",
			ast.Forbid().PrincipalEq(types.NewEntityUID("T", "42")),
			&Env{
				Principal: Ignore(),
			},
			ast.Forbid(),
			true,
		},
		{"ignoreAnd",
			ast.Permit().When(ast.Context().Access("variable").And(ast.Context().Access("ignore").Equal(ast.Long(42)))),
			&Env{
				Context: types.Record{
					"ignore":   Ignore(),
					"variable": Variable("variable"),
				},
			},
			ast.Permit(),
			true,
		},
		{"ignoreOr",
			ast.Permit().When(ast.Context().Access("variable").Or(ast.Context().Access("ignore").Equal(ast.Long(42)))),
			&Env{
				Context: types.Record{
					"ignore":   Ignore(),
					"variable": Variable("variable"),
				},
			},
			ast.Permit(),
			true,
		},
		{"ignoreIfThen",
			ast.Permit().When(ast.IfThenElse(ast.Context().Access("variable"), ast.Context().Access("ignore").Equal(ast.Long(42)), ast.True())),
			&Env{
				Context: types.Record{
					"ignore":   Ignore(),
					"variable": Variable("variable"),
				},
			},
			ast.Permit(),
			true,
		},
		{"ignoreIfElse",
			ast.Permit().When(ast.IfThenElse(ast.Context().Access("variable"), ast.True(), ast.Context().Access("ignore").Equal(ast.Long(42)))),
			&Env{
				Context: types.Record{
					"ignore":   Ignore(),
					"variable": Variable("variable"),
				},
			},
			ast.Permit(),
			true,
		},
		{"ignoreHas",
			ast.Permit().When(ast.Context().Has("ignore")),
			&Env{
				Context: types.Record{
					"ignore":   Ignore(),
					"variable": Variable("variable"),
				},
			},
			ast.Permit(),
			true,
		},
		{"ignoreHasNot",
			ast.Permit().When(ast.Not(ast.Context().Has("ignore"))),
			&Env{
				Context: types.Record{
					"ignore":   Ignore(),
					"variable": Variable("variable"),
				},
			},
			ast.Permit(),
			true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, keep, _ := PartialPolicy(InitEnv(tt.env), tt.in)
			if keep {
				testutil.Equals(t, out, tt.out)
				// gotP := (*parser.Policy)(out)
				// wantP := (*parser.Policy)(tt.out)
				// var gotB bytes.Buffer
				// gotP.MarshalCedar(&gotB)
				// var wantB bytes.Buffer
				// wantP.MarshalCedar(&wantB)
				// testutil.Equals(t, gotB.String(), wantB.String())
			}
			testutil.Equals(t, keep, tt.keep)

		})
	}

}

func TestPartialIfThenElse(t *testing.T) {
	errorN := ast.Long(42).GreaterThan(ast.String("bananas"))
	trueN := ast.True()
	falseN := ast.False()
	valueN := ast.String("test")
	keepN := ast.Context()
	_, _, _, _, _ = errorN, trueN, falseN, valueN, keepN
	valueA := ast.String("a")
	valueB := ast.String("b")

	tests := []struct {
		name    string
		in      ast.Node
		out     any
		errTest func(testutil.TB, error)
	}{
		{"ifTrueAB", ast.IfThenElse(trueN, valueA, valueB), valueA, testutil.OK},
		{"ifFalseAB", ast.IfThenElse(falseN, valueA, valueB), valueB, testutil.OK},
		{"ifValueAB", ast.IfThenElse(valueN, valueA, valueB), nil, testutil.Error},
		{"ifKeepAB", ast.IfThenElse(keepN, valueA, valueB), ast.IfThenElse(keepN, valueA, valueB), testutil.OK},
		{"ifErrorAB", ast.IfThenElse(errorN, valueA, valueB), nil, testutil.Error},

		{"ifTrueErrorB", ast.IfThenElse(trueN, errorN, valueB), nil, testutil.Error},
		{"ifFalseAError", ast.IfThenElse(falseN, valueA, errorN), nil, testutil.Error},
		{"ifTrueAError", ast.IfThenElse(trueN, valueA, errorN), valueA, testutil.OK},
		{"ifFalseErrorB", ast.IfThenElse(falseN, errorN, valueB), valueB, testutil.OK},

		{"ifKeepKeepKeep", ast.IfThenElse(keepN, keepN, keepN), ast.IfThenElse(keepN, keepN, keepN), testutil.OK},
		{"ifKeepErrorError", ast.IfThenElse(keepN, errorN, errorN), ast.IfThenElse(keepN, ast.ExtensionCall("error"), ast.ExtensionCall("error")), testutil.OK},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n, ok := tt.in.AsIsNode().(ast.NodeTypeIfThenElse)
			testutil.Equals(t, ok, true)
			out, err := partialIfThenElse(&Env{
				Context: Variable("context"),
			}, n)
			tt.errTest(t, err)
			if err != nil {
				return
			}
			nd, ok := tt.out.(ast.Node)
			testutil.Equals(t, ok, true)
			testutil.Equals(t, out, nd.AsIsNode())
		})
	}
}

func TestPartialAnd(t *testing.T) {
	errorN := ast.Long(42).GreaterThan(ast.String("bananas"))
	trueN := ast.True()
	falseN := ast.False()
	valueN := ast.String("test")
	keepN := ast.Context()
	_, _, _, _, _ = errorN, trueN, falseN, valueN, keepN

	tests := []struct {
		name    string
		in      ast.Node
		out     any
		errTest func(testutil.TB, error)
	}{

		{"andTrueTrue", trueN.And(trueN), ast.True(), testutil.OK},
		{"andTrueFalse", trueN.And(falseN), ast.False(), testutil.OK},
		{"andTrueValue", trueN.And(valueN), nil, testutil.Error},
		{"andTrueKeep", trueN.And(keepN), trueN.And(keepN), testutil.OK},
		{"andTrueError", trueN.And(errorN), nil, testutil.Error},

		{"andFalseTrue", falseN.And(trueN), ast.False(), testutil.OK},
		{"andFalseFalse", falseN.And(falseN), ast.False(), testutil.OK},
		{"andFalseValue", falseN.And(valueN), ast.False(), testutil.OK},
		{"andFalseKeep", falseN.And(keepN), ast.False(), testutil.OK},
		{"andFalseError", falseN.And(errorN), ast.False(), testutil.OK},

		{"andValueTrue", valueN.And(trueN), nil, testutil.Error},
		{"andValueFalse", valueN.And(falseN), nil, testutil.Error},
		{"andValueValue", valueN.And(valueN), nil, testutil.Error},
		{"andValueKeep", valueN.And(keepN), nil, testutil.Error},
		{"andValueError", valueN.And(errorN), nil, testutil.Error},

		{"andKeepTrue", keepN.And(trueN), keepN.And(trueN), testutil.OK},
		{"andKeepFalse", keepN.And(falseN), keepN.And(falseN), testutil.OK},
		{"andKeepValue", keepN.And(valueN), keepN.And(valueN), testutil.OK},
		{"andKeepKeep", keepN.And(keepN), keepN.And(keepN), testutil.OK},
		{"andKeepError", keepN.And(errorN), keepN.And(ast.ExtensionCall("error")), testutil.OK},

		{"andErrorTrue", errorN.And(trueN), nil, testutil.Error},
		{"andErrorFalse", errorN.And(falseN), nil, testutil.Error},
		{"andErrorValue", errorN.And(valueN), nil, testutil.Error},
		{"andErrorKeep", errorN.And(keepN), nil, testutil.Error},
		{"andErrorError", errorN.And(errorN), nil, testutil.Error},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n, ok := tt.in.AsIsNode().(ast.NodeTypeAnd)
			testutil.Equals(t, ok, true)
			out, err := partialAnd(&Env{
				Context: Variable("context"),
			}, n)
			tt.errTest(t, err)
			if err != nil {
				return
			}
			nd, ok := tt.out.(ast.Node)
			testutil.Equals(t, ok, true)
			testutil.Equals(t, out, nd.AsIsNode())
		})
	}
}

func TestPartialOr(t *testing.T) {
	errorN := ast.Long(42).GreaterThan(ast.String("bananas"))
	trueN := ast.True()
	falseN := ast.False()
	valueN := ast.String("test")
	keepN := ast.Context()
	_, _, _, _, _ = errorN, trueN, falseN, valueN, keepN

	tests := []struct {
		name    string
		in      ast.Node
		out     any
		errTest func(testutil.TB, error)
	}{

		{"orTrueTrue", trueN.Or(trueN), ast.True(), testutil.OK},
		{"orTrueFalse", trueN.Or(falseN), ast.True(), testutil.OK},
		{"orTrueValue", trueN.Or(valueN), ast.True(), testutil.OK},
		{"orTrueKeep", trueN.Or(keepN), ast.True(), testutil.OK},
		{"orTrueError", trueN.Or(errorN), ast.True(), testutil.OK},

		{"orFalseTrue", falseN.Or(trueN), ast.True(), testutil.OK},
		{"orFalseFalse", falseN.Or(falseN), ast.False(), testutil.OK},
		{"orFalseValue", falseN.Or(valueN), nil, testutil.Error},
		{"orFalseKeep", falseN.Or(keepN), falseN.Or(keepN), testutil.OK},
		{"orFalseError", falseN.Or(errorN), nil, testutil.Error},

		{"orValueTrue", valueN.Or(trueN), nil, testutil.Error},
		{"orValueFalse", valueN.Or(falseN), nil, testutil.Error},
		{"orValueValue", valueN.Or(valueN), nil, testutil.Error},
		{"orValueKeep", valueN.Or(keepN), nil, testutil.Error},
		{"orValueError", valueN.Or(errorN), nil, testutil.Error},

		{"orKeepTrue", keepN.Or(trueN), keepN.Or(trueN), testutil.OK},
		{"orKeepFalse", keepN.Or(falseN), keepN.Or(falseN), testutil.OK},
		{"orKeepValue", keepN.Or(valueN), keepN.Or(valueN), testutil.OK},
		{"orKeepKeep", keepN.Or(keepN), keepN.Or(keepN), testutil.OK},
		{"orKeepError", keepN.Or(errorN), keepN.Or(ast.ExtensionCall("error")), testutil.OK},

		{"orErrorTrue", errorN.Or(trueN), nil, testutil.Error},
		{"orErrorFalse", errorN.Or(falseN), nil, testutil.Error},
		{"orErrorValue", errorN.Or(valueN), nil, testutil.Error},
		{"orErrorKeep", errorN.Or(keepN), nil, testutil.Error},
		{"orErrorError", errorN.Or(errorN), nil, testutil.Error},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n, ok := tt.in.AsIsNode().(ast.NodeTypeOr)
			testutil.Equals(t, ok, true)
			out, err := partialOr(&Env{
				Context: Variable("context"),
			}, n)
			tt.errTest(t, err)
			if err != nil {
				return
			}
			nd, ok := tt.out.(ast.Node)
			testutil.Equals(t, ok, true)
			testutil.Equals(t, out, nd.AsIsNode())
		})
	}
}
