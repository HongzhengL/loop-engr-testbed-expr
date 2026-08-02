package runtime_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/expr-lang/expr/internal/testify/assert"
	"github.com/expr-lang/expr/internal/testify/require"

	"github.com/expr-lang/expr/vm/runtime"
)

// --- FetchMethod -------------------------------------------------------

type fetchMethodValueReceiver struct{}

func (fetchMethodValueReceiver) Bar() int { return 7 }

// fetchMethodPtrOnly declares a method with a pointer receiver only, so it
// belongs to the pointer's method set but not to the value's method set.
type fetchMethodPtrOnly struct{}

func (*fetchMethodPtrOnly) Greet() string { return "hi" }

// TestFetchMethod_ReturnsBoundMethodAtIndex covers:
//
//loop:behavior fetchmethod-fetchmethod-returns-the-bound-method-at-method-i
func TestFetchMethod_ReturnsBoundMethodAtIndex(t *testing.T) {
	from := fetchMethodValueReceiver{}
	typ := reflect.TypeOf(from)
	require.Equal(t, 1, typ.NumMethod(), "test fixture expects exactly one method")
	idx := typ.Method(0).Index
	name := typ.Method(0).Name

	got := runtime.FetchMethod(from, &runtime.Method{Index: idx, Name: name})
	fn, ok := got.(func() int)
	require.True(t, ok, "FetchMethod returned %T; want func() int", got)
	require.Equal(t, 7, fn(), "bound method call = %v; want 7", fn())
}

// fetchMethodTwoMethods declares two methods so tests can check FetchMethod
// dispatches strictly by the given Method.Index rather than, say, always
// resolving the first method regardless of index.
type fetchMethodTwoMethods struct{}

func (fetchMethodTwoMethods) Alpha() int { return 1 }
func (fetchMethodTwoMethods) Beta() int  { return 2 }

// TestFetchMethod_SelectsMethodByIndexNotJustTheFirstOne covers:
//
//loop:behavior fetchmethod-fetchmethod-returns-the-bound-method-at-method-i
func TestFetchMethod_SelectsMethodByIndexNotJustTheFirstOne(t *testing.T) {
	from := fetchMethodTwoMethods{}
	typ := reflect.TypeOf(from)
	require.Equal(t, 2, typ.NumMethod(), "test fixture expects exactly two methods")

	results := map[string]int{}
	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)
		got := runtime.FetchMethod(from, &runtime.Method{Index: m.Index, Name: m.Name})
		fn, ok := got.(func() int)
		require.True(t, ok, "FetchMethod(from, {Index:%d,Name:%q}) returned %T; want func() int", m.Index, m.Name, got)
		results[m.Name] = fn()
	}
	require.Equal(t, 1, results["Alpha"], "FetchMethod at Alpha's index should call Alpha, got %v", results)
	require.Equal(t, 2, results["Beta"], "FetchMethod at Beta's index should call Beta, got %v", results)
}

// TestFetchMethod_PanicsWhenFromIsInvalidUntypedNil covers:
//
//loop:behavior fetchmethod-fetchmethod-panics-when-from-is-an-invalid-untyp
func TestFetchMethod_PanicsWhenFromIsInvalidUntypedNil(t *testing.T) {
	method := &runtime.Method{Index: 0, Name: "X"}
	defer func() {
		r := recover()
		require.NotNil(t, r, "expected FetchMethod to panic, got none")
		msg := fmt.Sprint(r)
		assert.Contains(t, msg, "cannot fetch", "panic message = %q, want to contain %q", msg, "cannot fetch")
		assert.Contains(t, msg, method.Name, "panic message = %q, want to contain method name %q", msg, method.Name)
	}()
	runtime.FetchMethod(nil, method)
}

// TestFetchMethod_ResolvesPointerOnlyMethodsWithoutExplicitDereference covers:
//
//loop:behavior fetchmethod-fetchmethod-resolves-pointer-only-methods-withou
func TestFetchMethod_ResolvesPointerOnlyMethodsWithoutExplicitDereference(t *testing.T) {
	from := &fetchMethodPtrOnly{}
	typ := reflect.TypeOf(from)
	require.Equal(t, 1, typ.NumMethod(), "test fixture expects exactly one pointer-only method")
	idx := typ.Method(0).Index
	name := typ.Method(0).Name

	// Sanity check on the fixture: the value's method set (not the pointer's)
	// must not contain this method, so a successful call below demonstrates
	// FetchMethod resolved it without dereferencing from first.
	valueTyp := reflect.TypeOf(fetchMethodPtrOnly{})
	require.Equal(t, 0, valueTyp.NumMethod(), "test fixture expects the value method set to be empty")

	got := runtime.FetchMethod(from, &runtime.Method{Index: idx, Name: name})
	fn, ok := got.(func() string)
	require.True(t, ok, "FetchMethod returned %T; want func() string", got)
	require.Equal(t, "hi", fn(), "bound method call = %v; want %q", fn(), "hi")
}
