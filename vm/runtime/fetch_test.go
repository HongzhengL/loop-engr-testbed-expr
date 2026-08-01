package runtime_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/expr-lang/expr/internal/testify/assert"
	"github.com/expr-lang/expr/internal/testify/require"

	"github.com/expr-lang/expr/vm/runtime"
)

// --- fixtures shared by the tests in this file -----------------------------

type fetchSimpleStruct struct {
	Title string
}

type fetchAliasStruct struct {
	Name string `expr:"alias"`
}

type fetchDashStruct struct {
	Secret string `expr:"-"`
}

type fetchUnexportedStruct struct {
	secret string
}

// FetchNamer and FetchConcreteWithName must be exported: Go's reflect
// package treats an anonymous struct field as unexported when the field's
// type name (used as the field name for promotion purposes) is unexported,
// which would make the promoted field unreachable via reflection even
// though the field name itself ("Title") is exported.
type FetchNamer interface {
	Name() string
}

type FetchConcreteWithName struct {
	Title string
}

func (FetchConcreteWithName) Name() string { return "" }

type FetchEmbedInterface struct {
	FetchNamer
}

type fetchMethodStruct struct {
	Title string
}

func (fetchMethodStruct) Foo() int { return 42 }

// --- Fetch -------------------------------------------------------------

// TestFetch_ReturnsExportedStructFieldByName covers:
//
//loop:behavior fetch-fetch-returns-exported-struct-field-by-name
func TestFetch_ReturnsExportedStructFieldByName(t *testing.T) {
	from := fetchSimpleStruct{Title: "hello"}
	got := runtime.Fetch(from, "Title")
	require.Equal(t, "hello", got)
}

// TestFetch_ResolvesFieldViaExprTagAlias covers:
//
//loop:behavior fetch-fetch-resolves-field-via-expr-tag-alias
func TestFetch_ResolvesFieldViaExprTagAlias(t *testing.T) {
	from := fetchAliasStruct{Name: "bob"}
	got := runtime.Fetch(from, "alias")
	require.Equal(t, "bob", got)
}

// TestFetch_PanicsOnFieldTaggedExprDash covers:
//
//loop:behavior fetch-fetch-panics-on-field-tagged-expr-dash
func TestFetch_PanicsOnFieldTaggedExprDash(t *testing.T) {
	// Message text is not part of the contract for this behavior, only that
	// it panics.
	defer func() {
		r := recover()
		require.NotNil(t, r, "expected Fetch to panic, got none")
	}()
	runtime.Fetch(fetchDashStruct{Secret: "x"}, "Secret")
}

// TestFetch_PanicsOnUnexportedOrMissingField covers:
//
//loop:behavior fetch-fetch-panics-on-unexported-or-missing-field
func TestFetch_PanicsOnUnexportedOrMissingField(t *testing.T) {
	tests := []struct {
		name string
		from any
		i    any
	}{
		{"unexported field requested", fetchUnexportedStruct{secret: "x"}, "secret"},
		{"missing field", fetchSimpleStruct{Title: "x"}, "Missing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				require.NotNil(t, r, "expected Fetch to panic, got none")
				msg := fmt.Sprint(r)
				assert.Contains(t, msg, "cannot fetch", "panic message = %q, want to contain %q", msg, "cannot fetch")
			}()
			runtime.Fetch(tt.from, tt.i)
		})
	}
}

// TestFetch_DereferencesPointerToStructBeforeFieldLookup covers:
//
//loop:behavior fetch-fetch-dereferences-pointer-to-struct-before-fiel
func TestFetch_DereferencesPointerToStructBeforeFieldLookup(t *testing.T) {
	from := &fetchSimpleStruct{Title: "ptr-value"}
	got := runtime.Fetch(from, "Title")
	require.Equal(t, "ptr-value", got)
}

// TestFetch_PanicsOnInvalidNilInterfaceInput covers:
//
//loop:behavior fetch-fetch-panics-on-invalid-nil-interface-input
func TestFetch_PanicsOnInvalidNilInterfaceInput(t *testing.T) {
	defer func() {
		r := recover()
		require.NotNil(t, r, "expected Fetch to panic, got none")
		msg := fmt.Sprint(r)
		assert.Contains(t, msg, "cannot fetch", "panic message = %q, want to contain %q", msg, "cannot fetch")
	}()
	runtime.Fetch(nil, "x")
}

// TestFetch_ReturnsSliceOrArrayElementByPositiveIndex covers:
//
//loop:behavior fetch-fetch-returns-slice-or-array-element-by-positive
func TestFetch_ReturnsSliceOrArrayElementByPositiveIndex(t *testing.T) {
	tests := []struct {
		name string
		from any
		i    any
		want any
	}{
		{"slice, int index", []int{10, 20, 30}, 1, 20},
		{"array, last valid index", [3]string{"a", "b", "c"}, 2, "c"},
		{"slice, non-int numeric index type", []int{5, 6, 7}, uint8(2), 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.Fetch(tt.from, tt.i)
			require.Equal(t, tt.want, got)
		})
	}
}

// TestFetch_ReturnsSliceOrArrayElementByNegativeIndex covers:
//
//loop:behavior fetch-fetch-returns-slice-or-array-element-by-negative
func TestFetch_ReturnsSliceOrArrayElementByNegativeIndex(t *testing.T) {
	tests := []struct {
		name string
		from any
		i    any
		want any
	}{
		{"last element via -1", []int{10, 20, 30}, -1, 30},
		{"first element via -len (boundary)", []int{10, 20, 30}, -3, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.Fetch(tt.from, tt.i)
			require.Equal(t, tt.want, got)
		})
	}
}

// TestFetch_PanicsOnOutOfRangeIndex covers:
//
//loop:behavior fetch-fetch-panics-on-out-of-range-index
func TestFetch_PanicsOnOutOfRangeIndex(t *testing.T) {
	tests := []struct {
		name string
		from any
		i    any
	}{
		{"one past end (boundary)", []int{1, 2, 3}, 3},
		{"far past end", []int{1, 2, 3}, 10},
		{"one before start after wrap (boundary)", []int{1, 2, 3}, -4},
		{"far negative", []int{1, 2, 3}, -100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				require.NotNil(t, r, "expected Fetch to panic, got none")
				msg := fmt.Sprint(r)
				assert.Contains(t, msg, "index out of range", "panic message = %q, want to contain %q", msg, "index out of range")
			}()
			runtime.Fetch(tt.from, tt.i)
		})
	}
}

// TestFetch_PanicsWhenIndexingArrayOrSliceByStringKey covers:
//
//loop:behavior fetch-fetch-panics-when-indexing-array-or-slice-by-str
func TestFetch_PanicsWhenIndexingArrayOrSliceByStringKey(t *testing.T) {
	tests := []struct {
		name string
		from any
	}{
		{"slice", []int{1, 2, 3}},
		{"array", [3]int{1, 2, 3}},
		{"string", "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				require.NotNil(t, r, "expected Fetch to panic, got none")
				msg := fmt.Sprint(r)
				assert.Contains(t, msg, "cannot fetch", "panic message = %q, want to contain %q", msg, "cannot fetch")
			}()
			runtime.Fetch(tt.from, "x")
		})
	}
}

// TestFetch_ReturnsValueForPresentMapKey covers:
//
//loop:behavior fetch-fetch-returns-value-for-present-map-key
func TestFetch_ReturnsValueForPresentMapKey(t *testing.T) {
	from := map[string]int{"a": 1, "b": 2}
	got := runtime.Fetch(from, "b")
	require.Equal(t, 2, got)
}

// TestFetch_ReturnsZeroValueForAbsentMapKey covers:
//
//loop:behavior fetch-fetch-returns-zero-value-for-absent-map-key
func TestFetch_ReturnsZeroValueForAbsentMapKey(t *testing.T) {
	tests := []struct {
		name string
		from any
		i    any
		want any
	}{
		{"absent key, int values", map[string]int{"a": 1}, "z", 0},
		{"absent key, string values", map[string]string{"a": "x"}, "z", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.Fetch(tt.from, tt.i)
			require.Equal(t, tt.want, got)
		})
	}
}

// TestFetch_WithNilKeyLooksUpMapZeroValueKey covers:
//
//loop:behavior fetch-fetch-with-nil-key-looks-up-map-zero-value-key
func TestFetch_WithNilKeyLooksUpMapZeroValueKey(t *testing.T) {
	t.Run("zero-value key present", func(t *testing.T) {
		from := map[string]int{"": 99, "a": 1}
		got := runtime.Fetch(from, nil)
		require.Equal(t, 99, got)
	})
	t.Run("zero-value key absent", func(t *testing.T) {
		from := map[string]int{"a": 1}
		got := runtime.Fetch(from, nil)
		require.Equal(t, 0, got)
	})
}

// TestFetch_ResolvesMethodByNameString covers:
//
//loop:behavior fetch-fetch-resolves-method-by-name-string
func TestFetch_ResolvesMethodByNameString(t *testing.T) {
	from := fetchMethodStruct{Title: "x"}
	got := runtime.Fetch(from, "Foo")
	fn, ok := got.(func() int)
	require.True(t, ok, "expected Fetch to return a func() int, got %T", got)
	require.Equal(t, 42, fn())
}

// TestFetch_PanicsWhenNothingMatchesOnStructKind covers:
//
//loop:behavior fetch-fetch-panics-when-nothing-matches-on-struct-kind
func TestFetch_PanicsWhenNothingMatchesOnStructKind(t *testing.T) {
	// fetchMethodStruct has a method set (NumMethod() > 0 via Foo), but "Bar"
	// is neither a method name nor a field name.
	from := fetchMethodStruct{Title: "x"}
	defer func() {
		r := recover()
		require.NotNil(t, r, "expected Fetch to panic, got none")
		msg := fmt.Sprint(r)
		assert.Contains(t, msg, "cannot fetch", "panic message = %q, want to contain %q", msg, "cannot fetch")
	}()
	runtime.Fetch(from, "Bar")
}

// TestFetch_FindsFieldThroughEmbeddedInterfaceConcreteValue covers:
//
//loop:behavior fetch-fetch-finds-field-through-embedded-interface-s-c
func TestFetch_FindsFieldThroughEmbeddedInterfaceConcreteValue(t *testing.T) {
	// The concrete value is held via a pointer: reflect's rules around
	// unexported-field read-only propagation behave differently for values
	// boxed directly in an interface vs. reached through a pointer, so the
	// pointer form is used here to exercise the traversal itself rather
	// than an unrelated reflect quirk.
	from := FetchEmbedInterface{FetchNamer: &FetchConcreteWithName{Title: "embedded"}}
	got := runtime.Fetch(from, "Title")
	require.Equal(t, "embedded", got)
}

// --- FetchField ----------------------------------------------------------

type ffSingle struct {
	Title string
}

type ffInner struct {
	Value int
}

type ffOuter struct {
	Inner ffInner
}

type ffOuterPtr struct {
	Inner *ffInner
}

// TestFetchField_ReturnsValueForSingleIndexPath covers:
//
//loop:behavior fetchfield-fetchfield-returns-value-for-single-index-path
func TestFetchField_ReturnsValueForSingleIndexPath(t *testing.T) {
	field := &runtime.Field{Index: []int{0}, Path: []string{"Title"}}
	got := runtime.FetchField(ffSingle{Title: "hi"}, field)
	require.Equal(t, "hi", got)
}

// TestFetchField_WalksNestedStructPath covers:
//
//loop:behavior fetchfield-fetchfield-walks-nested-struct-path
func TestFetchField_WalksNestedStructPath(t *testing.T) {
	field := &runtime.Field{Index: []int{0, 0}, Path: []string{"Inner", "Value"}}

	t.Run("plain nested struct", func(t *testing.T) {
		from := ffOuter{Inner: ffInner{Value: 42}}
		got := runtime.FetchField(from, field)
		require.Equal(t, 42, got)
	})

	t.Run("nested struct through non-nil pointer", func(t *testing.T) {
		from := ffOuterPtr{Inner: &ffInner{Value: 7}}
		got := runtime.FetchField(from, field)
		require.Equal(t, 7, got)
	})
}

// TestFetchField_PanicsOnNilPointerMidPath covers:
//
//loop:behavior fetchfield-fetchfield-panics-on-nil-pointer-mid-path
func TestFetchField_PanicsOnNilPointerMidPath(t *testing.T) {
	field := &runtime.Field{Index: []int{0, 0}, Path: []string{"Inner", "Value"}}
	from := ffOuterPtr{Inner: nil}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic, got none")
		}
		msg := fmt.Sprint(r)
		if !strings.Contains(msg, "Inner") && !strings.Contains(msg, "Value") {
			t.Errorf("panic message = %q, want it to reference a field.Path segment (%q or %q)", msg, "Inner", "Value")
		}
	}()
	runtime.FetchField(from, field)
}

// TestFetchField_PanicsOnInvalidFrom covers:
//
//loop:behavior fetchfield-fetchfield-panics-on-invalid-from
func TestFetchField_PanicsOnInvalidFrom(t *testing.T) {
	field := &runtime.Field{Index: []int{0}, Path: []string{"Title"}}
	defer func() {
		r := recover()
		require.NotNil(t, r, "expected FetchField to panic, got none")
		msg := fmt.Sprint(r)
		assert.Contains(t, msg, "Title", "panic message = %q, want to contain %q", msg, "Title")
	}()
	runtime.FetchField(nil, field)
}
