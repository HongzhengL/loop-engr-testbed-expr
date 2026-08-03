package patcher_test

import (
	"reflect"
	"testing"

	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/builtin"
	"github.com/expr-lang/expr/checker/nature"
	"github.com/expr-lang/expr/conf"
	"github.com/expr-lang/expr/internal/testify/require"
	"github.com/expr-lang/expr/patcher"
)

// addEnv is used as an Env-resolvable overload: a method with the correct
// (receiver, int, int) -> int signature.
type addEnv struct{}

func (addEnv) Add(a, b int) int { return a + b }

// wrongArityMethodEnv exposes an Add method whose signature, once the
// receiver is accounted for, does not have exactly two logical inputs.
type wrongArityMethodEnv struct{}

func (wrongArityMethodEnv) Add(a int) int { return a }

// wrongOutputMethodEnv exposes an Add method whose signature has the right
// number of inputs but the wrong number of outputs.
type wrongOutputMethodEnv struct{}

func (wrongOutputMethodEnv) Add(a, b int) (int, int) { return a, b }

// newNatureEnv builds a *nature.Nature/*nature.Cache pair describing t, using
// a real (non-nil) cache so that both method and struct-field lookups work,
// mirroring how conf.EnvWithCache builds an OperatorOverloading.Env in
// production (see expr.Operator).
func newNatureEnv(t reflect.Type) (*nature.Nature, *nature.Cache) {
	cache := &nature.Cache{}
	n := cache.FromType(t)
	return &n, cache
}

// mustRecoverError runs f and requires that it panics with a non-nil error
// value, returning that error for further assertions.
func mustRecoverError(t *testing.T, f func()) error {
	t.Helper()
	var (
		err      error
		didPanic bool
	)
	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
				if e, ok := r.(error); ok {
					err = e
				} else {
					t.Fatalf("want panic value to be an error, got %T: %v", r, r)
				}
			}
		}()
		f()
	}()
	if !didPanic {
		t.Fatalf("want a panic, got none")
	}
	return err
}

func newIntIdent(name string) *ast.IdentifierNode {
	n := &ast.IdentifierNode{Value: name}
	n.SetType(reflect.TypeOf(0))
	return n
}

func TestOperatorOverloading_Visit_MatchingOperatorWithSuitableOverloadReplacesBinaryNodeWithCall(t *testing.T) {
	// loop:behavior operatoroverloading-visit-matching-operator-with-suitable-overload-replace
	left := newIntIdent("a")
	right := newIntIdent("b")
	bin := &ast.BinaryNode{Operator: "+", Left: left, Right: right}
	var node ast.Node = bin

	p := &patcher.OperatorOverloading{
		Operator:  "+",
		Overloads: []string{"Add"},
		Functions: conf.FunctionsTable{
			"Add": &builtin.Function{Types: []reflect.Type{reflect.TypeOf(func(int, int) int { return 0 })}},
		},
	}

	p.Visit(&node)

	call, ok := node.(*ast.CallNode)
	require.True(t, ok, "want node replaced with *ast.CallNode, got %T", node)

	callee, ok := call.Callee.(*ast.IdentifierNode)
	require.True(t, ok, "want Callee to be *ast.IdentifierNode, got %T", call.Callee)
	require.Equal(t, "Add", callee.Value)

	require.Equal(t, []ast.Node{left, right}, call.Arguments)
	require.Equal(t, reflect.TypeOf(0), call.Type())
}

func TestOperatorOverloading_Visit_BinaryNodeWithDifferentOperatorTokenIsLeftUntouched(t *testing.T) {
	// loop:behavior operatoroverloading-visit-binary-node-with-different-operator-token-is-lef
	left := newIntIdent("a")
	right := newIntIdent("b")
	bin := &ast.BinaryNode{Operator: "-", Left: left, Right: right}
	var node ast.Node = bin

	p := &patcher.OperatorOverloading{
		Operator:  "+",
		Overloads: []string{"Add"},
		Functions: conf.FunctionsTable{
			"Add": &builtin.Function{Types: []reflect.Type{reflect.TypeOf(func(int, int) int { return 0 })}},
		},
	}

	p.Visit(&node)

	require.Same(t, bin, node)
}

func TestOperatorOverloading_Visit_NonBinaryNodeIsLeftUntouched(t *testing.T) {
	// loop:behavior operatoroverloading-visit-non-binary-node-is-left-untouched
	ident := &ast.IdentifierNode{Value: "x"}
	var node ast.Node = ident

	p := &patcher.OperatorOverloading{Operator: "+"}
	p.Visit(&node)

	require.Same(t, ident, node)
}

func TestOperatorOverloading_Visit_MatchingOperatorWithNoSuitableOverloadLeavesNodeUntouched(t *testing.T) {
	// loop:behavior operatoroverloading-visit-matching-operator-with-no-suitable-overload-leav
	left := newIntIdent("a")
	right := newIntIdent("b")
	bin := &ast.BinaryNode{Operator: "+", Left: left, Right: right}
	var node ast.Node = bin

	env, cache := newNatureEnv(reflect.TypeOf(struct{}{}))
	p := &patcher.OperatorOverloading{
		Operator:  "+",
		Overloads: []string{"Add"},
		Functions: conf.FunctionsTable{
			// "Add" exists, but does not accept (int, int).
			"Add": &builtin.Function{Types: []reflect.Type{reflect.TypeOf(func(string, string) string { return "" })}},
		},
		Env:     env,
		NtCache: cache,
	}

	p.Visit(&node)

	require.Same(t, bin, node)
}

func TestOperatorOverloading_FindSuitableOperatorOverload_MatchFoundViaFunctionsTableOrEnvReturnsOverload(t *testing.T) {
	// loop:behavior operatoroverloading-findsuitable-match-found-via-functions-table-or-env-returns-o
	intType := reflect.TypeOf(0)

	t.Run("via functions table", func(t *testing.T) {
		env, cache := newNatureEnv(reflect.TypeOf(struct{}{}))
		p := &patcher.OperatorOverloading{
			Overloads: []string{"Add"},
			Functions: conf.FunctionsTable{
				"Add": &builtin.Function{Types: []reflect.Type{reflect.TypeOf(func(int, int) int { return 0 })}},
			},
			Env:     env,
			NtCache: cache,
		}

		ret, name, ok := p.FindSuitableOperatorOverload(intType, intType)
		require.True(t, ok)
		require.Equal(t, "Add", name)
		require.Equal(t, intType, ret)
	})

	t.Run("via env", func(t *testing.T) {
		env, cache := newNatureEnv(reflect.TypeOf(addEnv{}))
		p := &patcher.OperatorOverloading{
			Overloads: []string{"Add"},
			Env:       env,
			NtCache:   cache,
		}

		ret, name, ok := p.FindSuitableOperatorOverload(intType, intType)
		require.True(t, ok)
		require.Equal(t, "Add", name)
		require.Equal(t, intType, ret)
	})
}

func TestOperatorOverloading_FindSuitableOperatorOverload_NoCandidateAcceptsOperandTypesReturnsZeroValuesAndFalse(t *testing.T) {
	// loop:behavior operatoroverloading-findsuitable-no-candidate-accepts-operand-types-returns-zero
	intType := reflect.TypeOf(0)

	env, cache := newNatureEnv(reflect.TypeOf(addEnv{}))
	p := &patcher.OperatorOverloading{
		Overloads: []string{"AddFunc", "AddMethod"},
		Functions: conf.FunctionsTable{
			// present, but only accepts (string, string), not (int, int).
			"AddFunc": &builtin.Function{Types: []reflect.Type{reflect.TypeOf(func(string, string) string { return "" })}},
		},
		// "AddMethod" does not exist on addEnv either.
		Env:     env,
		NtCache: cache,
	}

	ret, name, ok := p.FindSuitableOperatorOverload(intType, intType)
	require.False(t, ok)
	require.Equal(t, "", name)
	require.Nil(t, ret)
}

func TestOperatorOverloading_Check_PanicsWhenAnOverloadNameResolvesToNeitherFunctionsNorEnv(t *testing.T) {
	// loop:behavior operatoroverloading-check-check-panics-when-an-overload-name-resolves-to-n
	t.Run("absent from both functions and env", func(t *testing.T) {
		env, cache := newNatureEnv(reflect.TypeOf(struct{}{}))
		p := &patcher.OperatorOverloading{
			Operator:  "+",
			Overloads: []string{"Add"},
			Env:       env,
			NtCache:   cache,
		}

		err := mustRecoverError(t, p.Check)
		require.Contains(t, err.Error(), "does not exist in the environment")
		require.Contains(t, err.Error(), "Add")
		require.Contains(t, err.Error(), "+")
	})

	t.Run("present in env but not of func kind", func(t *testing.T) {
		type envWithField struct {
			Add int
		}
		env, cache := newNatureEnv(reflect.TypeOf(envWithField{}))
		p := &patcher.OperatorOverloading{
			Operator:  "+",
			Overloads: []string{"Add"},
			Env:       env,
			NtCache:   cache,
		}

		err := mustRecoverError(t, p.Check)
		require.Contains(t, err.Error(), "does not exist in the environment")
	})

	t.Run("resolves normally: no panic", func(t *testing.T) {
		env, cache := newNatureEnv(reflect.TypeOf(struct{}{}))
		p := &patcher.OperatorOverloading{
			Operator:  "+",
			Overloads: []string{"Add"},
			Functions: conf.FunctionsTable{
				"Add": &builtin.Function{Types: []reflect.Type{reflect.TypeOf(func(int, int) int { return 0 })}},
			},
			Env:     env,
			NtCache: cache,
		}

		require.NotPanics(t, p.Check)
	})
}

func TestOperatorOverloading_Check_PanicsWhenAResolvedOverloadHasTheWrongArityOrReturnCount(t *testing.T) {
	// loop:behavior operatoroverloading-check-check-panics-when-a-resolved-overload-has-the-wr
	t.Run("functions table entry with too few inputs", func(t *testing.T) {
		env, cache := newNatureEnv(reflect.TypeOf(struct{}{}))
		p := &patcher.OperatorOverloading{
			Operator:  "+",
			Overloads: []string{"Add"},
			Functions: conf.FunctionsTable{
				"Add": &builtin.Function{Types: []reflect.Type{reflect.TypeOf(func(int) int { return 0 })}},
			},
			Env:     env,
			NtCache: cache,
		}

		err := mustRecoverError(t, p.Check)
		require.Contains(t, err.Error(), "correct signature")
		require.Contains(t, err.Error(), "Add")
	})

	t.Run("functions table entry with too many outputs", func(t *testing.T) {
		env, cache := newNatureEnv(reflect.TypeOf(struct{}{}))
		p := &patcher.OperatorOverloading{
			Operator:  "+",
			Overloads: []string{"Add"},
			Functions: conf.FunctionsTable{
				"Add": &builtin.Function{Types: []reflect.Type{reflect.TypeOf(func(int, int) (int, int) { return 0, 0 })}},
			},
			Env:     env,
			NtCache: cache,
		}

		err := mustRecoverError(t, p.Check)
		require.Contains(t, err.Error(), "correct signature")
	})

	t.Run("env method with wrong arity once receiver is accounted for", func(t *testing.T) {
		env, cache := newNatureEnv(reflect.TypeOf(wrongArityMethodEnv{}))
		p := &patcher.OperatorOverloading{
			Operator:  "+",
			Overloads: []string{"Add"},
			Env:       env,
			NtCache:   cache,
		}

		err := mustRecoverError(t, p.Check)
		require.Contains(t, err.Error(), "correct signature")
		require.Contains(t, err.Error(), "Add")
	})

	t.Run("env method with wrong output count", func(t *testing.T) {
		env, cache := newNatureEnv(reflect.TypeOf(wrongOutputMethodEnv{}))
		p := &patcher.OperatorOverloading{
			Operator:  "+",
			Overloads: []string{"Add"},
			Env:       env,
			NtCache:   cache,
		}

		err := mustRecoverError(t, p.Check)
		require.Contains(t, err.Error(), "correct signature")
	})
}

func TestOperatorOverloading_Reset_ClearsAppliedTrackingBeforeTheNextWalk(t *testing.T) {
	// loop:behavior operatoroverloading-reset-reset-clears-applied-tracking-before-the-next-wa
	left := newIntIdent("a")
	right := newIntIdent("b")
	bin := &ast.BinaryNode{Operator: "+", Left: left, Right: right}
	var node ast.Node = bin

	p := &patcher.OperatorOverloading{
		Operator:  "+",
		Overloads: []string{"Add"},
		Functions: conf.FunctionsTable{
			"Add": &builtin.Function{Types: []reflect.Type{reflect.TypeOf(func(int, int) int { return 0 })}},
		},
	}

	p.Visit(&node)
	require.True(t, p.ShouldRepeat(), "want ShouldRepeat true after a successful patch")

	p.Reset()
	require.False(t, p.ShouldRepeat(), "want ShouldRepeat false immediately after Reset")
}

func TestOperatorOverloading_ShouldRepeat_ReportsTrueOnlyAfterVisitAppliesAPatchSinceReset(t *testing.T) {
	// loop:behavior operatoroverloading-shouldrepeat-should-repeat-reports-true-only-after-visit-appl
	p := &patcher.OperatorOverloading{
		Operator:  "+",
		Overloads: []string{"Add"},
		Functions: conf.FunctionsTable{
			"Add": &builtin.Function{Types: []reflect.Type{reflect.TypeOf(func(int, int) int { return 0 })}},
		},
	}
	p.Reset()
	require.False(t, p.ShouldRepeat(), "want ShouldRepeat false right after Reset, before any Visit")

	// A Visit call that makes no replacement (operator mismatch) must not flip ShouldRepeat.
	noMatchLeft := newIntIdent("a")
	noMatchRight := newIntIdent("b")
	noMatch := &ast.BinaryNode{Operator: "-", Left: noMatchLeft, Right: noMatchRight}
	var noMatchNode ast.Node = noMatch
	p.Visit(&noMatchNode)
	require.False(t, p.ShouldRepeat(), "want ShouldRepeat false when Visit made no replacement")

	// A Visit call that does apply a replacement flips ShouldRepeat to true.
	matchLeft := newIntIdent("c")
	matchRight := newIntIdent("d")
	match := &ast.BinaryNode{Operator: "+", Left: matchLeft, Right: matchRight}
	var matchNode ast.Node = match
	p.Visit(&matchNode)
	require.True(t, p.ShouldRepeat(), "want ShouldRepeat true after Visit applies a replacement")
}
