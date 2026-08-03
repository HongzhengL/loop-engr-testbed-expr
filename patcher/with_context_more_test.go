package patcher_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/builtin"
	"github.com/expr-lang/expr/conf"
	"github.com/expr-lang/expr/internal/testify/require"
	"github.com/expr-lang/expr/patcher"
)

func TestWithContext_Visit_CallWithZeroParameterCalleeIsLeftUntouched(t *testing.T) {
	// loop:behavior withcontext-visit-call-with-zero-parameter-callee-is-left-untouche
	callee := &ast.IdentifierNode{Value: "fn"}
	callee.SetType(reflect.TypeOf(func() int { return 0 }))
	call := &ast.CallNode{Callee: callee}
	var node ast.Node = call

	w := patcher.WithContext{Name: "ctx"}
	w.Visit(&node)

	require.Same(t, call, node)
	require.Len(t, call.Arguments, 0)
}

func TestWithContext_Visit_NonCallNodeIsLeftUntouched(t *testing.T) {
	// loop:behavior withcontext-visit-non-call-node-is-left-untouched
	ident := &ast.IdentifierNode{Value: "x"}
	var node ast.Node = ident

	w := patcher.WithContext{Name: "ctx"}
	w.Visit(&node)

	require.Same(t, ident, node)
}

func TestWithContext_Visit_CallWhoseCalleeTypeCannotBeResolvedToAFuncIsLeftUntouched(t *testing.T) {
	// loop:behavior withcontext-visit-call-whose-callee-type-cannot-be-resolved-to-a-f
	callee := &ast.IdentifierNode{Value: "fn"} // no SetType call: Type() reports interface{} (unknown)
	call := &ast.CallNode{Callee: callee}
	var node ast.Node = call

	w := patcher.WithContext{Name: "ctx"} // Functions and Env left at their zero values (nil)
	w.Visit(&node)

	require.Same(t, call, node)
}

func TestWithContext_Visit_UnknownCalleeTypeResolvedThroughTheFunctionsTableBeforePatching(t *testing.T) {
	// loop:behavior withcontext-visit-unknown-callee-type-resolved-through-the-functio
	callee := &ast.IdentifierNode{Value: "fn"} // callee type unknown at this point
	arg := &ast.IntegerNode{Value: 5}
	arg.SetType(reflect.TypeOf(0))
	call := &ast.CallNode{Callee: callee, Arguments: []ast.Node{arg}}
	var node ast.Node = call

	functionsTable := conf.FunctionsTable{
		"fn": &builtin.Function{Types: []reflect.Type{reflect.TypeOf(func(context.Context, int) int { return 0 })}},
	}
	w := patcher.WithContext{Name: "ctx", Functions: functionsTable}
	w.Visit(&node)

	patched, ok := node.(*ast.CallNode)
	require.True(t, ok, "want node patched into *ast.CallNode, got %T", node)
	require.Same(t, callee, patched.Callee)
	require.Len(t, patched.Arguments, 2)

	ctxArg, ok := patched.Arguments[0].(*ast.IdentifierNode)
	require.True(t, ok, "want first argument to be *ast.IdentifierNode, got %T", patched.Arguments[0])
	require.Equal(t, "ctx", ctxArg.Value)

	require.Same(t, arg, patched.Arguments[1])
}
