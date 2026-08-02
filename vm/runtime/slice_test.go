package runtime_test

import (
	"fmt"
	"testing"

	"github.com/expr-lang/expr/internal/testify/assert"
	"github.com/expr-lang/expr/internal/testify/require"

	"github.com/expr-lang/expr/vm/runtime"
)

// --- Slice -----------------------------------------------------------------

// TestSlice_ReturnsSubSliceForInBoundsIndices covers:
//
//loop:behavior slice-slice-returns-a-sub-slice-for-in-bounds-indices
func TestSlice_ReturnsSubSliceForInBoundsIndices(t *testing.T) {
	tests := []struct {
		name       string
		array      []int
		from, to   int
	}{
		{"middle range", []int{0, 1, 2, 3, 4}, 1, 3},
		{"whole slice", []int{0, 1, 2, 3, 4}, 0, 5},
		{"empty range, from == to", []int{0, 1, 2, 3, 4}, 2, 2},
		{"from zero to len", []int{10, 20, 30}, 0, 3},
		{"single element", []int{10, 20, 30}, 1, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.Slice(tt.array, tt.from, tt.to)
			want := tt.array[tt.from:tt.to]
			require.Equal(t, want, got, "Slice(%v, %v, %v) = %v; want %v", tt.array, tt.from, tt.to, got, want)
		})
	}
}

// TestSlice_ReturnsSubstringForStringInput covers:
//
//loop:behavior slice-slice-returns-a-substring-for-a-string-input
func TestSlice_ReturnsSubstringForStringInput(t *testing.T) {
	tests := []struct {
		name     string
		array    string
		from, to int
	}{
		{"middle substring", "hello world", 2, 8},
		{"whole string", "hello", 0, 5},
		{"empty substring", "hello", 3, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.Slice(tt.array, tt.from, tt.to)
			want := tt.array[tt.from:tt.to]
			require.Equal(t, want, got, "Slice(%q, %v, %v) = %v; want %q", tt.array, tt.from, tt.to, got, want)
		})
	}
}

// TestSlice_TreatsNegativeFromAsOffsetFromLength covers:
//
//loop:behavior slice-slice-treats-a-negative-from-as-an-offset-from-t
func TestSlice_TreatsNegativeFromAsOffsetFromLength(t *testing.T) {
	array := []int{0, 1, 2, 3, 4}
	// N=5, from=-2 -> resolved start = 5 + (-2) = 3.
	got := runtime.Slice(array, -2, 5)
	want := runtime.Slice(array, 3, 5)
	require.Equal(t, want, got, "Slice(array, -2, 5) = %v; want same as Slice(array, 3, 5) = %v", got, want)
	require.Equal(t, array[3:5], got)
}

// TestSlice_ClampsStartToZeroWhenStillNegativeAfterOffset covers:
//
//loop:behavior slice-slice-clamps-the-start-index-to-zero-when-still
func TestSlice_ClampsStartToZeroWhenStillNegativeAfterOffset(t *testing.T) {
	array := []int{0, 1, 2}
	// N=3, from=-10 -> N+from = -7, still negative, so effective start is 0.
	got := runtime.Slice(array, -10, 2)
	want := array[0:2]
	require.Equal(t, want, got, "Slice(array, -10, 2) = %v; want %v", got, want)
}

// TestSlice_TreatsNegativeToAsOffsetFromLength covers:
//
//loop:behavior slice-slice-treats-a-negative-to-as-an-offset-from-the
func TestSlice_TreatsNegativeToAsOffsetFromLength(t *testing.T) {
	array := []int{0, 1, 2, 3, 4}
	// N=5, to=-1 -> resolved end = 5 + (-1) = 4.
	got := runtime.Slice(array, 0, -1)
	want := array[0:4]
	require.Equal(t, want, got, "Slice(array, 0, -1) = %v; want %v", got, want)
}

// TestSlice_ClampsEndIndexToLengthWhenExceedsIt covers:
//
//loop:behavior slice-slice-clamps-the-end-index-to-the-length-when-it
func TestSlice_ClampsEndIndexToLengthWhenExceedsIt(t *testing.T) {
	array := []int{0, 1, 2}
	got := runtime.Slice(array, 0, 100)
	want := array[0:3]
	require.Equal(t, want, got, "Slice(array, 0, 100) = %v; want %v", got, want)
}

// TestSlice_ClampsStartToComputedEndWhenStartExceedsEnd covers:
//
//loop:behavior slice-slice-clamps-start-to-the-computed-end-when-star
func TestSlice_ClampsStartToComputedEndWhenStartExceedsEnd(t *testing.T) {
	t.Run("slice", func(t *testing.T) {
		array := []int{0, 1, 2, 3, 4}
		got := runtime.Slice(array, 3, 1)
		gotSlice, ok := got.([]int)
		require.True(t, ok, "Slice(array, 3, 1) returned %T; want []int", got)
		require.Equal(t, 0, len(gotSlice), "Slice(array, 3, 1) = %v; want empty slice", gotSlice)
	})
	t.Run("string", func(t *testing.T) {
		s := "hello"
		got := runtime.Slice(s, 4, 1)
		gotStr, ok := got.(string)
		require.True(t, ok, "Slice(string, 4, 1) returned %T; want string", got)
		require.Equal(t, "", gotStr, "Slice(%q, 4, 1) = %q; want empty string", s, gotStr)
	})
}

// TestSlice_ResolvesEffectiveBoundsAcrossFromToCombinations covers:
//
// This exercises the documented index-resolution rules (negative-as-offset,
// zero/length clamping, start-clamped-to-end) across a broad cross product
// of from/to values against a fixed-length slice and string, rather than
// the one-combination-per-behavior spot checks above. want is computed
// independently from the documented rules (not from the implementation),
// so a mutation in any single clamping step is expected to surface as a
// mismatch on at least one combination in the matrix.
//
//loop:behavior slice-slice-returns-a-sub-slice-for-in-bounds-indices
//loop:behavior slice-slice-returns-a-substring-for-a-string-input
//loop:behavior slice-slice-treats-a-negative-from-as-an-offset-from-t
//loop:behavior slice-slice-clamps-the-start-index-to-zero-when-still
//loop:behavior slice-slice-treats-a-negative-to-as-an-offset-from-the
//loop:behavior slice-slice-clamps-the-end-index-to-the-length-when-it
//loop:behavior slice-slice-clamps-start-to-the-computed-end-when-star
func TestSlice_ResolvesEffectiveBoundsAcrossFromToCombinations(t *testing.T) {
	const length = 6

	// resolveStart implements the documented resolution rule for `from`:
	// negative values are treated as an offset from the length, then
	// clamped to zero if still negative.
	resolveStart := func(x int) int {
		if x < 0 {
			x = length + x
		}
		if x < 0 {
			x = 0
		}
		return x
	}
	// resolveEnd implements the documented resolution rule for `to` for the
	// values exercised here (all chosen so length+to >= 0, i.e. within the
	// documented "negative to as offset" case): negative values are treated
	// as an offset from the length, then clamped to the length if it
	// exceeds it.
	resolveEnd := func(x int) int {
		if x < 0 {
			x = length + x
		}
		if x > length {
			x = length
		}
		return x
	}

	froms := []int{0, 1, 3, 5, 6, -1, -3, -6}
	tos := []int{0, 1, 3, 5, 6, 100, -1, -3, -6}

	array := []int{0, 1, 2, 3, 4, 5}
	str := "abcdef"
	require.Equal(t, length, len(array), "test fixture: array length must match `length`")
	require.Equal(t, length, len(str), "test fixture: string length must match `length`")

	for _, from := range froms {
		for _, to := range tos {
			from, to := from, to
			a := resolveStart(from)
			b := resolveEnd(to)
			if a > b {
				a = b
			}

			t.Run(fmt.Sprintf("slice from=%d,to=%d", from, to), func(t *testing.T) {
				want := array[a:b]
				got := runtime.Slice(array, from, to)
				require.Equal(t, want, got, "Slice(array, %d, %d) = %v; want %v (resolved start=%d, end=%d)", from, to, got, want, a, b)
			})
			t.Run(fmt.Sprintf("string from=%d,to=%d", from, to), func(t *testing.T) {
				want := str[a:b]
				got := runtime.Slice(str, from, to)
				require.Equal(t, want, got, "Slice(%q, %d, %d) = %v; want %q (resolved start=%d, end=%d)", str, from, to, got, want, a, b)
			})
		}
	}
}

// TestSlice_DereferencesPointerToSliceOrArray covers:
//
//loop:behavior slice-slice-dereferences-a-pointer-to-a-slice-or-array
func TestSlice_DereferencesPointerToSliceOrArray(t *testing.T) {
	t.Run("pointer to slice", func(t *testing.T) {
		array := []int{0, 1, 2, 3, 4}
		got := runtime.Slice(&array, 1, 3)
		want := runtime.Slice(array, 1, 3)
		require.Equal(t, want, got, "Slice(&array, 1, 3) = %v; want same as Slice(array, 1, 3) = %v", got, want)
	})
	t.Run("pointer to array", func(t *testing.T) {
		array := [5]int{0, 1, 2, 3, 4}
		got := runtime.Slice(&array, 1, 3)
		want := runtime.Slice(array, 1, 3)
		require.Equal(t, want, got, "Slice(&array, 1, 3) = %v; want same as Slice(array, 1, 3) = %v", got, want)
	})
}

// TestSlice_PanicsForUnsupportedInputKind covers:
//
//loop:behavior slice-slice-panics-for-an-unsupported-input-kind
func TestSlice_PanicsForUnsupportedInputKind(t *testing.T) {
	tests := []struct {
		name  string
		array any
	}{
		{"int", 42},
		{"map", map[string]int{"a": 1}},
		{"bool", true},
		{"struct", struct{}{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				require.NotNil(t, r, "expected Slice to panic, got none")
				msg := fmt.Sprint(r)
				assert.Contains(t, msg, "cannot slice", "panic message = %q, want to contain %q", msg, "cannot slice")
			}()
			runtime.Slice(tt.array, 0, 1)
		})
	}
}
