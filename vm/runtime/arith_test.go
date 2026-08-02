package runtime_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/expr-lang/expr/internal/testify/assert"
	"github.com/expr-lang/expr/internal/testify/require"

	"github.com/expr-lang/expr/vm/runtime"
)

// --- Exponent ------------------------------------------------------------

// TestExponent_ReturnsMathPowOfBothOperandsConvertedToFloat64 covers:
//
//loop:behavior exponent-exponent-returns-math-pow-of-both-operands-conve
func TestExponent_ReturnsMathPowOfBothOperandsConvertedToFloat64(t *testing.T) {
	tests := []struct {
		name string
		a, b any
		want float64
	}{
		{"int, float64", int(2), float64(10), math.Pow(2, 10)},
		{"float32, int8", float32(2), int8(3), math.Pow(2, 3)},
		{"uint, uint", uint(3), uint(2), math.Pow(3, 2)},
		{"int64, int64, zero exponent", int64(5), int64(0), math.Pow(5, 0)},
		{"float64, float64, fractional base", float64(2.5), float64(2), math.Pow(2.5, 2)},
		{"zero base, positive exponent", float64(0), float64(3), math.Pow(0, 3)},
		{"negative exponent", float64(2), float64(-2), math.Pow(2, -2)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.Exponent(tt.a, tt.b)
			require.Equal(t, tt.want, got, "Exponent(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
		})
	}
}

// TestExponent_PanicsWhenEitherOperandIsNotSupportedNumericKind covers:
//
//loop:behavior exponent-exponent-panics-when-either-operand-is-not-a-sup
func TestExponent_PanicsWhenEitherOperandIsNotSupportedNumericKind(t *testing.T) {
	tests := []struct {
		name string
		a, b any
	}{
		{"a is string", "2", 3},
		{"b is string", 2, "3"},
		{"both non-numeric", "2", "3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				require.NotNil(t, r, "expected Exponent to panic, got none")
				msg := fmt.Sprint(r)
				assert.Contains(t, msg, "invalid operation", "panic message = %q, want to contain %q", msg, "invalid operation")
			}()
			runtime.Exponent(tt.a, tt.b)
		})
	}
}

// --- Add -------------------------------------------------------------------

// TestAdd_SumsTwoNumericOperandsWidenedToIntOrFloat64 covers:
//
//loop:behavior add-add-sums-two-numeric-operands-widened-to-int-or
func TestAdd_SumsTwoNumericOperandsWidenedToIntOrFloat64(t *testing.T) {
	tests := []struct {
		name     string
		a, b     any
		wantInt  int
		wantIsInt bool
		wantFlt  float64
	}{
		{"int + int", int(2), int(3), 5, true, 0},
		{"uint + int8", uint(2), int8(3), 5, true, 0},
		{"uint8 + uint16", uint8(2), uint16(3), 5, true, 0},
		{"int + float64", int(2), float64(3.5), 0, false, 5.5},
		{"float32 + int", float32(1.5), int(2), 0, false, 3.5},
		{"float64 + float64", float64(1.5), float64(2.5), 0, false, 4.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.Add(tt.a, tt.b)
			if tt.wantIsInt {
				v, ok := got.(int)
				require.True(t, ok, "Add(%v, %v) returned %T; want int", tt.a, tt.b, got)
				require.Equal(t, tt.wantInt, v, "Add(%v, %v) = %v; want %v", tt.a, tt.b, v, tt.wantInt)
			} else {
				v, ok := got.(float64)
				require.True(t, ok, "Add(%v, %v) returned %T; want float64", tt.a, tt.b, got)
				require.Equal(t, tt.wantFlt, v, "Add(%v, %v) = %v; want %v", tt.a, tt.b, v, tt.wantFlt)
			}
		})
	}
}

// addNumericKind describes one of the numeric kinds Add accepts, together
// with a constructor for a value of a given magnitude and whether the kind
// belongs to the float family (which determines the widening rule).
type addNumericKind struct {
	name    string
	value   func(v int) any
	isFloat bool
}

var addNumericKinds = []addNumericKind{
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

// TestAdd_SumsEveryPairOfSupportedNumericKinds covers:
//
// This exercises the full cross product of supported numeric kinds (as
// opposed to the handful of representative pairs above) so that the
// widening rule ("int when both operands are integer-family, float64 when
// either is a float") is verified for every combination, not just a sample.
//
//loop:behavior add-add-sums-two-numeric-operands-widened-to-int-or
func TestAdd_SumsEveryPairOfSupportedNumericKinds(t *testing.T) {
	const av, bv = 3, 4 // distinct, non-zero values so operand-order and
	// operator mistakes (e.g. subtraction instead of addition) are caught.
	for _, ka := range addNumericKinds {
		for _, kb := range addNumericKinds {
			ka, kb := ka, kb
			t.Run(ka.name+"+"+kb.name, func(t *testing.T) {
				a := ka.value(av)
				b := kb.value(bv)
				got := runtime.Add(a, b)
				if ka.isFloat || kb.isFloat {
					want := float64(av) + float64(bv)
					v, ok := got.(float64)
					require.True(t, ok, "Add(%v, %v) returned %T; want float64", a, b, got)
					require.Equal(t, want, v, "Add(%v, %v) = %v; want %v", a, b, v, want)
				} else {
					want := av + bv
					v, ok := got.(int)
					require.True(t, ok, "Add(%v, %v) returned %T; want int", a, b, got)
					require.Equal(t, want, v, "Add(%v, %v) = %v; want %v", a, b, v, want)
				}
			})
		}
	}
}

// TestAdd_ConcatenatesTwoStringOperands covers:
//
//loop:behavior add-add-concatenates-two-string-operands
func TestAdd_ConcatenatesTwoStringOperands(t *testing.T) {
	got := runtime.Add("foo", "bar")
	require.Equal(t, "foobar", got, `Add("foo", "bar") = %v; want "foobar"`, got)
}

// TestAdd_ReturnsLaterTimeWhenAddingDurationToTime covers:
//
//loop:behavior add-add-returns-a-later-time-time-when-adding-a-time
func TestAdd_ReturnsLaterTimeWhenAddingDurationToTime(t *testing.T) {
	base := time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)
	d := 2 * time.Hour
	got := runtime.Add(base, d)
	want := base.Add(d)
	require.Equal(t, want, got, "Add(base, d) = %v; want %v", got, want)
}

// TestAdd_IsDefinedForDurationPlusTimeInEitherOperandOrder covers:
//
//loop:behavior add-add-is-defined-for-time-duration-plus-time-time
func TestAdd_IsDefinedForDurationPlusTimeInEitherOperandOrder(t *testing.T) {
	base := time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)
	d := 90 * time.Minute

	gotDurationFirst := runtime.Add(d, base)
	gotTimeFirst := runtime.Add(base, d)

	require.Equal(t, gotTimeFirst, gotDurationFirst, "Add(d, base) = %v; want same as Add(base, d) = %v", gotDurationFirst, gotTimeFirst)
	require.Equal(t, base.Add(d), gotDurationFirst, "Add(d, base) = %v; want %v", gotDurationFirst, base.Add(d))
}

// TestAdd_SumsTwoTimeDurationOperands covers:
//
//loop:behavior add-add-sums-two-time-duration-operands
func TestAdd_SumsTwoTimeDurationOperands(t *testing.T) {
	d1 := 2 * time.Hour
	d2 := 30 * time.Minute
	got := runtime.Add(d1, d2)
	want := d1 + d2
	require.Equal(t, want, got, "Add(d1, d2) = %v; want %v", got, want)
}

// TestAdd_PanicsForOperandTypeCombinationsWithNoDefinedSum covers:
//
//loop:behavior add-add-panics-for-operand-type-combinations-with-no
func TestAdd_PanicsForOperandTypeCombinationsWithNoDefinedSum(t *testing.T) {
	tests := []struct {
		name string
		a, b any
	}{
		{"string + int", "foo", 1},
		{"bool + bool", true, false},
		{"int + bool", 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				require.NotNil(t, r, "expected Add to panic, got none")
				msg := fmt.Sprint(r)
				assert.Contains(t, msg, "invalid operation", "panic message = %q, want to contain %q", msg, "invalid operation")
			}()
			runtime.Add(tt.a, tt.b)
		})
	}
}
