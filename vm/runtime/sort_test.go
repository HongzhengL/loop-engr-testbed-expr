package runtime_test

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/expr-lang/expr/internal/testify/assert"
	"github.com/expr-lang/expr/internal/testify/require"

	"github.com/expr-lang/expr/vm/runtime"
)

// --- SortBy.Len --------------------------------------------------------

// TestSortByLen_ReturnsLengthOfArray covers:
//
//loop:behavior sortby-len-sortby-len-returns-length-of-array
func TestSortByLen_ReturnsLengthOfArray(t *testing.T) {
	tests := []struct {
		name   string
		array  []any
		values []any
	}{
		{"single element", []any{1}, []any{1}},
		{"multiple elements", []any{1, 2, 3, 4}, []any{4, 3, 2, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &runtime.SortBy{Array: tt.array, Values: tt.values}
			got := s.Len()
			require.Equal(t, len(tt.array), got, "SortBy.Len() = %v; want %v", got, len(tt.array))
		})
	}
}

// --- SortBy.Swap ---------------------------------------------------------

// TestSortBySwap_SwapsArrayAndValuesInLockstep covers:
//
//loop:behavior sortby-swap-sortby-swap-swaps-array-and-values-in-lockstep
func TestSortBySwap_SwapsArrayAndValuesInLockstep(t *testing.T) {
	s := &runtime.SortBy{
		Array:  []any{"a", "b", "c"},
		Values: []any{1, 2, 3},
	}
	beforeArrayI, beforeArrayJ := s.Array[0], s.Array[2]
	beforeValuesI, beforeValuesJ := s.Values[0], s.Values[2]

	s.Swap(0, 2)

	require.Equal(t, beforeArrayJ, s.Array[0], "Array[0] after Swap(0,2) = %v; want %v", s.Array[0], beforeArrayJ)
	require.Equal(t, beforeArrayI, s.Array[2], "Array[2] after Swap(0,2) = %v; want %v", s.Array[2], beforeArrayI)
	require.Equal(t, beforeValuesJ, s.Values[0], "Values[0] after Swap(0,2) = %v; want %v", s.Values[0], beforeValuesJ)
	require.Equal(t, beforeValuesI, s.Values[2], "Values[2] after Swap(0,2) = %v; want %v", s.Values[2], beforeValuesI)

	// pairing between an Array element and its corresponding Values element
	// is preserved: element that was at index 0 (Array="a", Values=1) is now
	// found together at index 2, and vice versa.
	require.Equal(t, "a", s.Array[2], "pairing broken: Array[2] = %v; want %v", s.Array[2], "a")
	require.Equal(t, 1, s.Values[2], "pairing broken: Values[2] = %v; want %v", s.Values[2], 1)
	require.Equal(t, "c", s.Array[0], "pairing broken: Array[0] = %v; want %v", s.Array[0], "c")
	require.Equal(t, 3, s.Values[0], "pairing broken: Values[0] = %v; want %v", s.Values[0], 3)
}

// --- SortBy.Less -----------------------------------------------------------

// TestSortByLess_AscendingComparesValuesViaPackageLess covers:
//
//loop:behavior sortby-less-sortby-less-ascending-compares-values-via-packag
func TestSortByLess_AscendingComparesValuesViaPackageLess(t *testing.T) {
	tests := []struct {
		name   string
		values []any
		i, j   int
	}{
		{"i < j", []any{1, 2}, 0, 1},
		{"i > j", []any{2, 1}, 0, 1},
		{"i == j value", []any{5, 5}, 0, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &runtime.SortBy{Desc: false, Array: []any{"x", "y"}, Values: tt.values}
			got := s.Less(tt.i, tt.j)
			want := runtime.Less(tt.values[tt.i], tt.values[tt.j])
			require.Equal(t, want, got, "SortBy{Desc:false}.Less(%d,%d) = %v; want Less(Values[%d],Values[%d]) = %v", tt.i, tt.j, got, tt.i, tt.j, want)
		})
	}
}

// TestSortByLess_DescendingReversesComparisonViaPackageLess covers:
//
//loop:behavior sortby-less-sortby-less-descending-reverses-the-comparison-v
func TestSortByLess_DescendingReversesComparisonViaPackageLess(t *testing.T) {
	tests := []struct {
		name   string
		values []any
		i, j   int
	}{
		{"i < j", []any{1, 2}, 0, 1},
		{"i > j", []any{2, 1}, 0, 1},
		{"i == j value", []any{5, 5}, 0, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &runtime.SortBy{Desc: true, Array: []any{"x", "y"}, Values: tt.values}
			got := s.Less(tt.i, tt.j)
			want := runtime.Less(tt.values[tt.j], tt.values[tt.i])
			require.Equal(t, want, got, "SortBy{Desc:true}.Less(%d,%d) = %v; want Less(Values[%d],Values[%d]) = %v", tt.i, tt.j, got, tt.j, tt.i, want)
		})
	}
}

// --- sort.Sort on SortBy ----------------------------------------------------

// TestSortBySort_ReordersArrayToMatchAscendingValuesOrder covers:
//
//loop:behavior sortby-sort-sort-on-sortby-reorders-array-to-match-asce
func TestSortBySort_ReordersArrayToMatchAscendingValuesOrder(t *testing.T) {
	// Array holds a label identifying which original pairing it came from,
	// so after sorting we can verify the Array/Values pairing was preserved.
	s := &runtime.SortBy{
		Desc:   false,
		Array:  []any{"label-30", "label-10", "label-20"},
		Values: []any{30, 10, 20},
	}
	pairing := map[any]any{
		"label-30": 30,
		"label-10": 10,
		"label-20": 20,
	}

	sort.Sort(s)

	require.True(t, sort.SliceIsSorted(s.Values, func(i, j int) bool {
		return runtime.Less(s.Values[i], s.Values[j])
	}), "Values after ascending sort.Sort = %v; want ascending order", s.Values)

	for i := range s.Array {
		want := pairing[s.Array[i]]
		require.Equal(t, want, s.Values[i], "pairing broken at index %d: Array=%v paired with Values=%v; want %v", i, s.Array[i], s.Values[i], want)
	}
}

// TestSortBySort_ReordersArrayToMatchDescendingValuesOrder covers:
//
//loop:behavior sortby-sort-sort-on-sortby-reorders-array-to-match-desc
func TestSortBySort_ReordersArrayToMatchDescendingValuesOrder(t *testing.T) {
	s := &runtime.SortBy{
		Desc:   true,
		Array:  []any{"label-30", "label-10", "label-20"},
		Values: []any{30, 10, 20},
	}
	pairing := map[any]any{
		"label-30": 30,
		"label-10": 10,
		"label-20": 20,
	}

	sort.Sort(s)

	require.True(t, sort.SliceIsSorted(s.Values, func(i, j int) bool {
		return runtime.Less(s.Values[j], s.Values[i])
	}), "Values after descending sort.Sort = %v; want descending order", s.Values)

	for i := range s.Array {
		want := pairing[s.Array[i]]
		require.Equal(t, want, s.Values[i], "pairing broken at index %d: Array=%v paired with Values=%v; want %v", i, s.Array[i], s.Values[i], want)
	}
}

// --- Sort.Len ----------------------------------------------------------

// TestSortLen_ReturnsLengthOfArray covers:
//
//loop:behavior sort-len-sort-len-returns-length-of-array
func TestSortLen_ReturnsLengthOfArray(t *testing.T) {
	tests := []struct {
		name  string
		array []any
	}{
		{"single element", []any{1}},
		{"multiple elements", []any{1, 2, 3, 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &runtime.Sort{Array: tt.array}
			got := s.Len()
			require.Equal(t, len(tt.array), got, "Sort.Len() = %v; want %v", got, len(tt.array))
		})
	}
}

// --- Sort.Swap -----------------------------------------------------------

// TestSortSwap_SwapsArrayElements covers:
//
//loop:behavior sort-swap-sort-swap-swaps-array-elements
func TestSortSwap_SwapsArrayElements(t *testing.T) {
	s := &runtime.Sort{Array: []any{"a", "b", "c"}}
	beforeI, beforeJ := s.Array[0], s.Array[2]

	s.Swap(0, 2)

	require.Equal(t, beforeJ, s.Array[0], "Array[0] after Swap(0,2) = %v; want %v", s.Array[0], beforeJ)
	require.Equal(t, beforeI, s.Array[2], "Array[2] after Swap(0,2) = %v; want %v", s.Array[2], beforeI)
}

// --- Sort.Less -------------------------------------------------------------

// TestSortLess_AscendingComparesArrayElementsViaPackageLess covers:
//
//loop:behavior sort-less-sort-less-ascending-compares-array-elements-via
func TestSortLess_AscendingComparesArrayElementsViaPackageLess(t *testing.T) {
	tests := []struct {
		name  string
		array []any
		i, j  int
	}{
		{"i < j", []any{1, 2}, 0, 1},
		{"i > j", []any{2, 1}, 0, 1},
		{"i == j value", []any{5, 5}, 0, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &runtime.Sort{Desc: false, Array: tt.array}
			got := s.Less(tt.i, tt.j)
			want := runtime.Less(tt.array[tt.i], tt.array[tt.j])
			require.Equal(t, want, got, "Sort{Desc:false}.Less(%d,%d) = %v; want Less(Array[%d],Array[%d]) = %v", tt.i, tt.j, got, tt.i, tt.j, want)
		})
	}
}

// TestSortLess_DescendingReversesComparisonViaPackageLess covers:
//
//loop:behavior sort-less-sort-less-descending-reverses-the-comparison-via
func TestSortLess_DescendingReversesComparisonViaPackageLess(t *testing.T) {
	tests := []struct {
		name  string
		array []any
		i, j  int
	}{
		{"i < j", []any{1, 2}, 0, 1},
		{"i > j", []any{2, 1}, 0, 1},
		{"i == j value", []any{5, 5}, 0, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &runtime.Sort{Desc: true, Array: tt.array}
			got := s.Less(tt.i, tt.j)
			want := runtime.Less(tt.array[tt.j], tt.array[tt.i])
			require.Equal(t, want, got, "Sort{Desc:true}.Less(%d,%d) = %v; want Less(Array[%d],Array[%d]) = %v", tt.i, tt.j, got, tt.j, tt.i, want)
		})
	}
}

// --- sort.Sort on Sort -----------------------------------------------------

// TestSortSort_SortsArrayAscendingInPlace covers:
//
//loop:behavior sort-sort-sort-on-sort-value-sorts-array-ascending-in
func TestSortSort_SortsArrayAscendingInPlace(t *testing.T) {
	s := &runtime.Sort{Desc: false, Array: []any{30, 10, 20, 10}}

	sort.Sort(s)

	require.True(t, sort.SliceIsSorted(s.Array, func(i, j int) bool {
		return runtime.Less(s.Array[i], s.Array[j])
	}), "Array after ascending sort.Sort = %v; want ascending order", s.Array)
}

// TestSortSort_SortsArrayDescendingInPlace covers:
//
//loop:behavior sort-sort-sort-on-sort-value-sorts-array-descending-i
func TestSortSort_SortsArrayDescendingInPlace(t *testing.T) {
	s := &runtime.Sort{Desc: true, Array: []any{30, 10, 20, 10}}

	sort.Sort(s)

	require.True(t, sort.SliceIsSorted(s.Array, func(i, j int) bool {
		return runtime.Less(s.Array[j], s.Array[i])
	}), "Array after descending sort.Sort = %v; want descending order", s.Array)
}

// --- Less --------------------------------------------------------------

// numericKind describes one of the numeric kinds Less/More accept, together
// with a constructor for a value of a given magnitude and whether the kind
// belongs to the float family (which determines the widening rule).
type numericKind struct {
	name    string
	value   func(v int) any
	isFloat bool
}

var lessNumericKinds = []numericKind{
	{"uint", func(v int) any { return uint(v) }, false},
	{"uint8", func(v int) any { return uint8(v) }, false},
	{"uint16", func(v int) any { return uint16(v) }, false},
	{"uint32", func(v int) any { return uint32(v) }, false},
	{"uint64", func(v int) any { return uint64(v) }, false},
	{"int", func(v int) any { return int(v) }, false},
	{"int8", func(v int) any { return int8(v) }, false},
	{"int16", func(v int) any { return int16(v) }, false},
	{"int32", func(v int) any { return int32(v) }, false},
	{"int64", func(v int) any { return int64(v) }, false},
	{"float32", func(v int) any { return float32(v) }, true},
	{"float64", func(v int) any { return float64(v) }, true},
}

// TestLess_ComparesNumericOperandsWidenedToIntOrFloat64 covers:
//
//loop:behavior less-less-compares-numeric-operands-widened-to-int-or
func TestLess_ComparesNumericOperandsWidenedToIntOrFloat64(t *testing.T) {
	// Cross product of every supported numeric kind, at both a smaller and
	// a larger magnitude, so the widening rule is verified in both
	// directions and not just for a handful of representative pairs.
	for _, ka := range lessNumericKinds {
		for _, kb := range lessNumericKinds {
			ka, kb := ka, kb
			t.Run(ka.name+"<"+kb.name, func(t *testing.T) {
				a := ka.value(3)
				b := kb.value(6)
				got := runtime.Less(a, b)
				var want bool
				if ka.isFloat || kb.isFloat {
					want = float64(3) < float64(6)
				} else {
					want = 3 < 6
				}
				require.Equal(t, want, got, "Less(%v, %v) = %v; want %v", a, b, got, want)

				// reversed operand order
				gotRev := runtime.Less(b, a)
				var wantRev bool
				if ka.isFloat || kb.isFloat {
					wantRev = float64(6) < float64(3)
				} else {
					wantRev = 6 < 3
				}
				require.Equal(t, wantRev, gotRev, "Less(%v, %v) = %v; want %v", b, a, gotRev, wantRev)
			})
		}
	}

	// Signed-integer sign boundary: negative vs zero, and two negatives.
	signTests := []struct {
		name string
		a, b any
		want bool
	}{
		{"negative < zero", int(-1), int(0), true},
		{"zero < negative", int(0), int(-1), false},
		{"negative < negative, a smaller", int(-5), int(-2), true},
		{"negative < negative, a larger", int(-2), int(-5), false},
		{"negative float < zero float", float64(-1.5), float64(0), true},
		{"negative int < positive float", int(-3), float64(2.5), true},
	}
	for _, tt := range signTests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.Less(tt.a, tt.b)
			require.Equal(t, tt.want, got, "Less(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
		})
	}
}

// TestLess_ComparesTwoStringOperandsLexicographically covers:
//
//loop:behavior less-less-compares-two-string-operands-lexicographica
func TestLess_ComparesTwoStringOperandsLexicographically(t *testing.T) {
	tests := []struct {
		name string
		a, b string
	}{
		{"a < b", "apple", "banana"},
		{"a > b", "banana", "apple"},
		{"a == b", "same", "same"},
		{"empty < non-empty", "", "a"},
		{"prefix < longer", "ab", "abc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.Less(tt.a, tt.b)
			want := tt.a < tt.b
			require.Equal(t, want, got, "Less(%q, %q) = %v; want %v", tt.a, tt.b, got, want)
		})
	}
}

// TestLess_ComparesTwoTimeTimeOperandsChronologically covers:
//
//loop:behavior less-less-compares-two-time-time-operands-chronologic
func TestLess_ComparesTwoTimeTimeOperandsChronologically(t *testing.T) {
	earlier := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	later := time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		a, b time.Time
	}{
		{"earlier < later", earlier, later},
		{"later < earlier", later, earlier},
		{"equal times", earlier, earlier},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.Less(tt.a, tt.b)
			want := tt.a.Before(tt.b)
			require.Equal(t, want, got, "Less(%v, %v) = %v; want %v", tt.a, tt.b, got, want)
		})
	}
}

// TestLess_ComparesTwoTimeDurationOperandsNumerically covers:
//
//loop:behavior less-less-compares-two-time-duration-operands-numeric
func TestLess_ComparesTwoTimeDurationOperandsNumerically(t *testing.T) {
	tests := []struct {
		name string
		a, b time.Duration
	}{
		{"shorter < longer", time.Second, time.Minute},
		{"longer < shorter", time.Minute, time.Second},
		{"equal durations", time.Second, time.Second},
		{"negative < positive", -time.Second, time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.Less(tt.a, tt.b)
			want := tt.a < tt.b
			require.Equal(t, want, got, "Less(%v, %v) = %v; want %v", tt.a, tt.b, got, want)
		})
	}
}

// TestLess_PanicsForOperandCombinationsWithNoDefinedOrder covers:
//
//loop:behavior less-less-panics-for-operand-combinations-with-no-def
func TestLess_PanicsForOperandCombinationsWithNoDefinedOrder(t *testing.T) {
	tests := []struct {
		name string
		a, b any
	}{
		{"string and int", "foo", 1},
		{"bool and bool", true, false},
		{"int and bool", 1, true},
		{"time.Time and string", time.Now(), "foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				require.NotNil(t, r, "expected Less(%v, %v) to panic, got none", tt.a, tt.b)
				msg := fmt.Sprint(r)
				assert.Contains(t, msg, "invalid operation", "panic message = %q, want to contain %q", msg, "invalid operation")
			}()
			runtime.Less(tt.a, tt.b)
		})
	}
}

// --- More ----------------------------------------------------------------

// TestMore_ComparesNumericOperandsWidenedToIntOrFloat64 covers:
//
//loop:behavior more-more-compares-numeric-operands-widened-to-int-or
func TestMore_ComparesNumericOperandsWidenedToIntOrFloat64(t *testing.T) {
	for _, ka := range lessNumericKinds {
		for _, kb := range lessNumericKinds {
			ka, kb := ka, kb
			t.Run(ka.name+">"+kb.name, func(t *testing.T) {
				a := ka.value(6)
				b := kb.value(3)
				got := runtime.More(a, b)
				var want bool
				if ka.isFloat || kb.isFloat {
					want = float64(6) > float64(3)
				} else {
					want = 6 > 3
				}
				require.Equal(t, want, got, "More(%v, %v) = %v; want %v", a, b, got, want)

				gotRev := runtime.More(b, a)
				var wantRev bool
				if ka.isFloat || kb.isFloat {
					wantRev = float64(3) > float64(6)
				} else {
					wantRev = 3 > 6
				}
				require.Equal(t, wantRev, gotRev, "More(%v, %v) = %v; want %v", b, a, gotRev, wantRev)
			})
		}
	}

	signTests := []struct {
		name string
		a, b any
		want bool
	}{
		{"zero > negative", int(0), int(-1), true},
		{"negative > zero", int(-1), int(0), false},
		{"negative > negative, a larger", int(-2), int(-5), true},
		{"negative > negative, a smaller", int(-5), int(-2), false},
		{"positive float > negative float", float64(1.5), float64(-1.5), true},
	}
	for _, tt := range signTests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.More(tt.a, tt.b)
			require.Equal(t, tt.want, got, "More(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
		})
	}
}

// TestMore_ComparesTwoStringOperandsLexicographically covers:
//
//loop:behavior more-more-compares-two-string-operands-lexicographica
func TestMore_ComparesTwoStringOperandsLexicographically(t *testing.T) {
	tests := []struct {
		name string
		a, b string
	}{
		{"a > b", "banana", "apple"},
		{"a < b", "apple", "banana"},
		{"a == b", "same", "same"},
		{"non-empty > empty", "a", ""},
		{"longer > prefix", "abc", "ab"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.More(tt.a, tt.b)
			want := tt.a > tt.b
			require.Equal(t, want, got, "More(%q, %q) = %v; want %v", tt.a, tt.b, got, want)
		})
	}
}

// TestMore_ComparesTwoTimeTimeOperandsChronologically covers:
//
//loop:behavior more-more-compares-two-time-time-operands-chronologic
func TestMore_ComparesTwoTimeTimeOperandsChronologically(t *testing.T) {
	earlier := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	later := time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		a, b time.Time
	}{
		{"later > earlier", later, earlier},
		{"earlier > later", earlier, later},
		{"equal times", earlier, earlier},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.More(tt.a, tt.b)
			want := tt.a.After(tt.b)
			require.Equal(t, want, got, "More(%v, %v) = %v; want %v", tt.a, tt.b, got, want)
		})
	}
}
