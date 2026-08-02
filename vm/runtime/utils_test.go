package runtime_test

import (
	"fmt"
	"testing"

	"github.com/expr-lang/expr/internal/testify/assert"
	"github.com/expr-lang/expr/internal/testify/require"

	"github.com/expr-lang/expr/vm/runtime"
)

// --- ToInt -----------------------------------------------------------------

// TestToInt_ConvertsEverySupportedNumericKind covers:
//
//loop:behavior toint-toint-converts-every-supported-numeric-kind-to-i
func TestToInt_ConvertsEverySupportedNumericKind(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want int
	}{
		{"float32, truncates toward zero", float32(3.9), 3},
		{"float64, truncates toward zero", float64(-3.9), -3},
		{"int", int(7), 7},
		{"int8", int8(-8), -8},
		{"int16", int16(1000), 1000},
		{"int32", int32(-70000), -70000},
		{"int64, value fits int on every supported platform", int64(1234567890), 1234567890},
		{"uint", uint(9), 9},
		{"uint8", uint8(255), 255},
		{"uint16", uint16(65000), 65000},
		{"uint32, value fits int on every supported platform", uint32(2000000000), 2000000000},
		{"uint64", uint64(42), 42},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.ToInt(tt.in)
			require.Equal(t, tt.want, got, "ToInt(%v) = %v; want %v", tt.in, got, tt.want)
		})
	}
}

// TestToInt_PanicsForNonNumericArgument covers:
//
//loop:behavior toint-toint-panics-for-a-non-numeric-argument
func TestToInt_PanicsForNonNumericArgument(t *testing.T) {
	tests := []struct {
		name string
		in   any
	}{
		{"string", "not a number"},
		{"bool", true},
		{"nil", nil},
		{"struct", struct{}{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				require.NotNil(t, r, "expected ToInt to panic, got none")
				msg := fmt.Sprint(r)
				assert.Contains(t, msg, "invalid operation", "panic message = %q, want to contain %q", msg, "invalid operation")
			}()
			runtime.ToInt(tt.in)
		})
	}
}

// --- ToInt64 -----------------------------------------------------------------

// TestToInt64_ConvertsEverySupportedNumericKind covers:
//
//loop:behavior toint64-toint64-converts-every-supported-numeric-kind-to
func TestToInt64_ConvertsEverySupportedNumericKind(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want int64
	}{
		{"float32, truncates toward zero", float32(3.9), int64(3)},
		{"float64, truncates toward zero", float64(-3.9), int64(-3)},
		{"int", int(7), int64(7)},
		{"int8", int8(-8), int64(-8)},
		{"int16", int16(1000), int64(1000)},
		{"int32", int32(-70000), int64(-70000)},
		{"int64", int64(9000000000), int64(9000000000)},
		{"uint", uint(9), int64(9)},
		{"uint8", uint8(255), int64(255)},
		{"uint16", uint16(65000), int64(65000)},
		{"uint32", uint32(4000000000), int64(4000000000)},
		{"uint64", uint64(42), int64(42)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.ToInt64(tt.in)
			require.Equal(t, tt.want, got, "ToInt64(%v) = %v; want %v", tt.in, got, tt.want)
		})
	}
}

// TestToInt64_PanicsForNonNumericArgument covers:
//
//loop:behavior toint64-toint64-panics-for-a-non-numeric-argument
func TestToInt64_PanicsForNonNumericArgument(t *testing.T) {
	tests := []struct {
		name string
		in   any
	}{
		{"string", "not a number"},
		{"bool", false},
		{"nil", nil},
		{"struct", struct{}{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				require.NotNil(t, r, "expected ToInt64 to panic, got none")
				msg := fmt.Sprint(r)
				assert.Contains(t, msg, "invalid operation", "panic message = %q, want to contain %q", msg, "invalid operation")
			}()
			runtime.ToInt64(tt.in)
		})
	}
}

// --- ToFloat64 -----------------------------------------------------------------

// TestToFloat64_ConvertsEverySupportedNumericKind covers:
//
//loop:behavior tofloat64-tofloat64-converts-every-supported-numeric-kind
func TestToFloat64_ConvertsEverySupportedNumericKind(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want float64
	}{
		{"float32", float32(3.5), float64(3.5)},
		{"float64", float64(-3.25), float64(-3.25)},
		{"int", int(7), float64(7)},
		{"int8", int8(-8), float64(-8)},
		{"int16", int16(1000), float64(1000)},
		{"int32", int32(-70000), float64(-70000)},
		{"int64", int64(9000000000), float64(9000000000)},
		{"uint", uint(9), float64(9)},
		{"uint8", uint8(255), float64(255)},
		{"uint16", uint16(65000), float64(65000)},
		{"uint32", uint32(4000000000), float64(4000000000)},
		{"uint64", uint64(42), float64(42)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.ToFloat64(tt.in)
			require.Equal(t, tt.want, got, "ToFloat64(%v) = %v; want %v", tt.in, got, tt.want)
		})
	}
}

// TestToFloat64_PanicsForNonNumericArgument covers:
//
//loop:behavior tofloat64-tofloat64-panics-for-a-non-numeric-argument
func TestToFloat64_PanicsForNonNumericArgument(t *testing.T) {
	tests := []struct {
		name string
		in   any
	}{
		{"string", "not a number"},
		{"bool", true},
		{"nil", nil},
		{"struct", struct{}{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				require.NotNil(t, r, "expected ToFloat64 to panic, got none")
				msg := fmt.Sprint(r)
				assert.Contains(t, msg, "invalid operation", "panic message = %q, want to contain %q", msg, "invalid operation")
			}()
			runtime.ToFloat64(tt.in)
		})
	}
}

// --- ToBool -----------------------------------------------------------------

// TestToBool_ReturnsFalseForNilArgument covers:
//
//loop:behavior tobool-tobool-returns-false-for-a-nil-argument
func TestToBool_ReturnsFalseForNilArgument(t *testing.T) {
	got := runtime.ToBool(nil)
	require.Equal(t, false, got, "ToBool(nil) = %v; want false", got)
}

// TestToBool_ReturnsUnderlyingBoolValue covers:
//
//loop:behavior tobool-tobool-returns-the-underlying-bool-value
func TestToBool_ReturnsUnderlyingBoolValue(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want bool
	}{
		{"true", true, true},
		{"false", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.ToBool(tt.in)
			require.Equal(t, tt.want, got, "ToBool(%v) = %v; want %v", tt.in, got, tt.want)
		})
	}
}

// TestToBool_PanicsForNonNilNonBoolArgument covers:
//
//loop:behavior tobool-tobool-panics-for-a-non-nil-non-bool-argument
func TestToBool_PanicsForNonNilNonBoolArgument(t *testing.T) {
	tests := []struct {
		name string
		in   any
	}{
		{"int", 1},
		{"string", "true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				require.NotNil(t, r, "expected ToBool to panic, got none")
				msg := fmt.Sprint(r)
				assert.Contains(t, msg, "invalid operation", "panic message = %q, want to contain %q", msg, "invalid operation")
			}()
			runtime.ToBool(tt.in)
		})
	}
}

// --- IsNil -------------------------------------------------------------

// TestIsNil_ReturnsTrueForNilInterfaceValue covers:
//
//loop:behavior isnil-isnil-returns-true-for-a-nil-interface-value
func TestIsNil_ReturnsTrueForNilInterfaceValue(t *testing.T) {
	got := runtime.IsNil(nil)
	require.True(t, got, "IsNil(nil) = %v; want true", got)
}

// TestIsNil_ReturnsTrueForNilPointerSliceMapChanOrFunc covers:
//
//loop:behavior isnil-isnil-returns-true-for-a-nil-pointer-slice-map-c
func TestIsNil_ReturnsTrueForNilPointerSliceMapChanOrFunc(t *testing.T) {
	var (
		nilPtr   *int
		nilSlice []int
		nilMap   map[string]int
		nilChan  chan int
		nilFunc  func()
	)
	tests := []struct {
		name string
		in   any
	}{
		{"nil pointer", nilPtr},
		{"nil slice", nilSlice},
		{"nil map", nilMap},
		{"nil chan", nilChan},
		{"nil func", nilFunc},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.IsNil(tt.in)
			require.True(t, got, "IsNil(%v) = %v; want true", tt.in, got)
		})
	}
}

// TestIsNil_ReturnsFalseForNonNilableKindEvenWhenZeroValued covers:
//
//loop:behavior isnil-isnil-returns-false-for-a-non-nilable-kind-even
func TestIsNil_ReturnsFalseForNonNilableKindEvenWhenZeroValued(t *testing.T) {
	tests := []struct {
		name string
		in   any
	}{
		{"zero int", int(0)},
		{"empty string", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.IsNil(tt.in)
			require.False(t, got, "IsNil(%v) = %v; want false", tt.in, got)
		})
	}
}

// --- Len -----------------------------------------------------------------

// TestLen_ReturnsElementCountForArraySliceMapAndString covers:
//
//loop:behavior len-len-returns-the-element-count-for-array-slice-ma
func TestLen_ReturnsElementCountForArraySliceMapAndString(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want int
	}{
		{"non-empty array", [3]int{1, 2, 3}, 3},
		{"empty array", [0]int{}, 0},
		{"non-empty slice", []int{1, 2, 3, 4}, 4},
		{"empty slice", []int{}, 0},
		{"nil slice", []int(nil), 0},
		{"non-empty map", map[string]int{"a": 1, "b": 2}, 2},
		{"empty map", map[string]int{}, 0},
		{"non-empty string", "hello", 5},
		{"empty string", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.Len(tt.in)
			require.Equal(t, tt.want, got, "Len(%v) = %v; want %v", tt.in, got, tt.want)
		})
	}
}

// TestLen_PanicsForTypeWithoutALength covers:
//
//loop:behavior len-len-panics-for-a-type-without-a-length
func TestLen_PanicsForTypeWithoutALength(t *testing.T) {
	tests := []struct {
		name string
		in   any
	}{
		{"int", 42},
		{"struct", struct{}{}},
		{"nil", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				require.NotNil(t, r, "expected Len to panic, got none")
				msg := fmt.Sprint(r)
				assert.Contains(t, msg, "invalid argument for len", "panic message = %q, want to contain %q", msg, "invalid argument for len")
			}()
			runtime.Len(tt.in)
		})
	}
}

// --- Negate ----------------------------------------------------------------

// TestNegate_ReturnsArithmeticNegationForEverySupportedNumericKind covers:
//
//loop:behavior negate-negate-returns-the-arithmetic-negation-for-every
func TestNegate_ReturnsArithmeticNegationForEverySupportedNumericKind(t *testing.T) {
	t.Run("float32", func(t *testing.T) {
		got := runtime.Negate(float32(3.5))
		v, ok := got.(float32)
		require.True(t, ok, "Negate(float32) returned %T; want float32", got)
		require.Equal(t, float32(-3.5), v)
	})
	t.Run("float64", func(t *testing.T) {
		got := runtime.Negate(float64(3.5))
		v, ok := got.(float64)
		require.True(t, ok, "Negate(float64) returned %T; want float64", got)
		require.Equal(t, float64(-3.5), v)
	})
	t.Run("int", func(t *testing.T) {
		got := runtime.Negate(int(7))
		v, ok := got.(int)
		require.True(t, ok, "Negate(int) returned %T; want int", got)
		require.Equal(t, int(-7), v)
	})
	t.Run("int8", func(t *testing.T) {
		got := runtime.Negate(int8(7))
		v, ok := got.(int8)
		require.True(t, ok, "Negate(int8) returned %T; want int8", got)
		require.Equal(t, int8(-7), v)
	})
	t.Run("int16", func(t *testing.T) {
		got := runtime.Negate(int16(7))
		v, ok := got.(int16)
		require.True(t, ok, "Negate(int16) returned %T; want int16", got)
		require.Equal(t, int16(-7), v)
	})
	t.Run("int32", func(t *testing.T) {
		got := runtime.Negate(int32(7))
		v, ok := got.(int32)
		require.True(t, ok, "Negate(int32) returned %T; want int32", got)
		require.Equal(t, int32(-7), v)
	})
	t.Run("int64", func(t *testing.T) {
		got := runtime.Negate(int64(7))
		v, ok := got.(int64)
		require.True(t, ok, "Negate(int64) returned %T; want int64", got)
		require.Equal(t, int64(-7), v)
	})
	t.Run("uint", func(t *testing.T) {
		got := runtime.Negate(uint(7))
		v, ok := got.(uint)
		require.True(t, ok, "Negate(uint) returned %T; want uint", got)
		var operand uint = 7
		want := -operand // computed at runtime: Go wraps unsigned negation, unlike a typed constant expression
		require.Equal(t, want, v)
	})
	t.Run("uint8", func(t *testing.T) {
		got := runtime.Negate(uint8(7))
		v, ok := got.(uint8)
		require.True(t, ok, "Negate(uint8) returned %T; want uint8", got)
		var operand uint8 = 7
		want := -operand
		require.Equal(t, want, v)
	})
	t.Run("uint16", func(t *testing.T) {
		got := runtime.Negate(uint16(7))
		v, ok := got.(uint16)
		require.True(t, ok, "Negate(uint16) returned %T; want uint16", got)
		var operand uint16 = 7
		want := -operand
		require.Equal(t, want, v)
	})
	t.Run("uint32", func(t *testing.T) {
		got := runtime.Negate(uint32(7))
		v, ok := got.(uint32)
		require.True(t, ok, "Negate(uint32) returned %T; want uint32", got)
		var operand uint32 = 7
		want := -operand
		require.Equal(t, want, v)
	})
	t.Run("uint64", func(t *testing.T) {
		got := runtime.Negate(uint64(7))
		v, ok := got.(uint64)
		require.True(t, ok, "Negate(uint64) returned %T; want uint64", got)
		var operand uint64 = 7
		want := -operand
		require.Equal(t, want, v)
	})
}

// TestNegate_PanicsForNonNumericArgument covers:
//
//loop:behavior negate-negate-panics-for-a-non-numeric-argument
func TestNegate_PanicsForNonNumericArgument(t *testing.T) {
	tests := []struct {
		name string
		in   any
	}{
		{"string", "7"},
		{"bool", true},
		{"nil", nil},
		{"struct", struct{}{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				require.NotNil(t, r, "expected Negate to panic, got none")
				msg := fmt.Sprint(r)
				assert.Contains(t, msg, "invalid operation", "panic message = %q, want to contain %q", msg, "invalid operation")
			}()
			runtime.Negate(tt.in)
		})
	}
}

// --- MakeRange ---------------------------------------------------------

// TestMakeRange_ReturnsInclusiveIntegerRangeFromMinToMax covers:
//
//loop:behavior makerange-makerange-returns-the-inclusive-integer-range-fr
func TestMakeRange_ReturnsInclusiveIntegerRangeFromMinToMax(t *testing.T) {
	tests := []struct {
		name     string
		min, max int
		want     []int
	}{
		{"1..5", 1, 5, []int{1, 2, 3, 4, 5}},
		{"negative to positive", -2, 2, []int{-2, -1, 0, 1, 2}},
		{"0..0 is single element", 0, 0, []int{0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.MakeRange(tt.min, tt.max)
			require.Equal(t, tt.want, got, "MakeRange(%v, %v) = %v; want %v", tt.min, tt.max, got, tt.want)
			require.Equal(t, tt.max-tt.min+1, len(got), "MakeRange(%v, %v) length = %v; want %v", tt.min, tt.max, len(got), tt.max-tt.min+1)
		})
	}
}

// TestMakeRange_ReturnsEmptySliceWhenMaxIsLessThanMin covers:
//
//loop:behavior makerange-makerange-returns-an-empty-slice-when-max-is-les
func TestMakeRange_ReturnsEmptySliceWhenMaxIsLessThanMin(t *testing.T) {
	tests := []struct {
		name     string
		min, max int
	}{
		{"max one below min", 5, 4},
		{"max well below min", 5, 1},
		{"negative max, positive min", 5, -5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.MakeRange(tt.min, tt.max)
			require.Equal(t, 0, len(got), "MakeRange(%v, %v) length = %v; want 0", tt.min, tt.max, len(got))
		})
	}
}

// TestMakeRange_ReturnsSingleElementSliceWhenMinEqualsMax covers:
//
//loop:behavior makerange-makerange-returns-a-single-element-slice-when-mi
func TestMakeRange_ReturnsSingleElementSliceWhenMinEqualsMax(t *testing.T) {
	tests := []struct {
		name string
		v    int
	}{
		{"zero", 0},
		{"positive", 42},
		{"negative", -42},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.MakeRange(tt.v, tt.v)
			require.Equal(t, []int{tt.v}, got, "MakeRange(%v, %v) = %v; want %v", tt.v, tt.v, got, []int{tt.v})
		})
	}
}

// --- In ----------------------------------------------------------------

// TestIn_ReportsWhetherNeedleEqualsElementOfSliceOrArray covers:
//
//loop:behavior in-in-reports-whether-needle-equals-an-element-of-a
func TestIn_ReportsWhetherNeedleEqualsElementOfSliceOrArray(t *testing.T) {
	tests := []struct {
		name   string
		needle any
		array  any
	}{
		{"present in slice", 2, []int{1, 2, 3}},
		{"present in array", "b", [3]string{"a", "b", "c"}},
		{"present via Equal-widened numeric comparison", int64(2), []int{1, 2, 3}},
		{"present as first element", 1, []int{1, 2, 3}},
		{"present as last element", 3, []int{1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.In(tt.needle, tt.array)
			require.True(t, got, "In(%v, %v) = %v; want true", tt.needle, tt.array, got)
		})
	}
}
