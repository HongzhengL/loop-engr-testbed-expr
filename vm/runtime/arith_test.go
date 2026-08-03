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
//loop:behavior exponent-exponent-returns-math-pow-of-both-operands-as-fl
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
//loop:behavior exponent-exponent-panics-when-either-operand-is-not-numer
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
//loop:behavior add-add-returns-a-later-time-value-when-adding-a-dur
func TestAdd_ReturnsLaterTimeWhenAddingDurationToTime(t *testing.T) {
	base := time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)
	d := 2 * time.Hour
	got := runtime.Add(base, d)
	want := base.Add(d)
	require.Equal(t, want, got, "Add(base, d) = %v; want %v", got, want)
}

// TestAdd_IsDefinedForDurationPlusTimeInEitherOperandOrder covers:
//
//loop:behavior add-add-is-defined-for-duration-plus-time-in-either
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
//loop:behavior add-add-panics-for-operand-combinations-with-no-defi
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

// --- Subtract --------------------------------------------------------------

// TestSubtract_WidensIntegerFamilyOperandsToIntAndFloatOperandsToFloat64 covers:
//
//loop:behavior subtract-subtract-widens-integer-family-operands-to-int-a
func TestSubtract_WidensIntegerFamilyOperandsToIntAndFloatOperandsToFloat64(t *testing.T) {
	tests := []struct {
		name    string
		a, b    any
		wantInt int
		isInt   bool
		wantFlt float64
	}{
		{"int - int", int(5), int(3), 2, true, 0},
		{"uint - int8", uint(5), int8(3), 2, true, 0},
		{"uint8 - uint16", uint8(9), uint16(4), 5, true, 0},
		{"int - float64", int(5), float64(3.5), 0, false, 1.5},
		{"float32 - int", float32(5.5), int(2), 0, false, 3.5},
		{"float64 - float64", float64(5.5), float64(2.5), 0, false, 3.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.Subtract(tt.a, tt.b)
			if tt.isInt {
				v, ok := got.(int)
				require.True(t, ok, "Subtract(%v, %v) returned %T; want int", tt.a, tt.b, got)
				require.Equal(t, tt.wantInt, v, "Subtract(%v, %v) = %v; want %v", tt.a, tt.b, v, tt.wantInt)
			} else {
				v, ok := got.(float64)
				require.True(t, ok, "Subtract(%v, %v) returned %T; want float64", tt.a, tt.b, got)
				require.Equal(t, tt.wantFlt, v, "Subtract(%v, %v) = %v; want %v", tt.a, tt.b, v, tt.wantFlt)
			}
		})
	}
}

// TestSubtract_SubtractsEveryPairOfSupportedNumericKinds exercises the full cross
// product of supported numeric kinds (mirroring TestAdd_SumsEveryPairOfSupportedNumericKinds)
// so the widening rule is verified for every combination, with operands of
// different magnitudes so a mis-ordered subtraction would be caught.
//
//loop:behavior subtract-subtract-widens-integer-family-operands-to-int-a
func TestSubtract_SubtractsEveryPairOfSupportedNumericKinds(t *testing.T) {
	const av, bv = 9, 4 // distinct values so operand order errors are caught
	for _, ka := range addNumericKinds {
		for _, kb := range addNumericKinds {
			ka, kb := ka, kb
			t.Run(ka.name+"-"+kb.name, func(t *testing.T) {
				a := ka.value(av)
				b := kb.value(bv)
				got := runtime.Subtract(a, b)
				if ka.isFloat || kb.isFloat {
					want := float64(av) - float64(bv)
					v, ok := got.(float64)
					require.True(t, ok, "Subtract(%v, %v) returned %T; want float64", a, b, got)
					require.Equal(t, want, v, "Subtract(%v, %v) = %v; want %v", a, b, v, want)
				} else {
					want := av - bv
					v, ok := got.(int)
					require.True(t, ok, "Subtract(%v, %v) returned %T; want int", a, b, got)
					require.Equal(t, want, v, "Subtract(%v, %v) = %v; want %v", a, b, v, want)
				}
			})
		}
	}
}

// TestSubtract_OfTwoTimeValuesReturnsTheirDurationDifference covers:
//
//loop:behavior subtract-subtract-of-two-time-values-returns-their-durati
func TestSubtract_OfTwoTimeValuesReturnsTheirDurationDifference(t *testing.T) {
	a := time.Date(2021, time.January, 1, 12, 0, 0, 0, time.UTC)
	b := time.Date(2021, time.January, 1, 10, 0, 0, 0, time.UTC)
	got := runtime.Subtract(a, b)
	want := a.Sub(b)
	require.Equal(t, want, got, "Subtract(a, b) = %v; want %v", got, want)
}

// TestSubtract_OfTimeMinusDurationReturnsAnEarlierTime covers:
//
//loop:behavior subtract-subtract-of-time-minus-duration-returns-an-earli
func TestSubtract_OfTimeMinusDurationReturnsAnEarlierTime(t *testing.T) {
	a := time.Date(2021, time.January, 1, 12, 0, 0, 0, time.UTC)
	d := 3 * time.Hour
	got := runtime.Subtract(a, d)
	want := a.Add(-d)
	require.Equal(t, want, got, "Subtract(a, d) = %v; want %v", got, want)
}

// TestSubtract_OfTwoDurationValuesReturnsTheirDifference covers:
//
//loop:behavior subtract-subtract-of-two-duration-values-returns-their-di
func TestSubtract_OfTwoDurationValuesReturnsTheirDifference(t *testing.T) {
	d1 := 3 * time.Hour
	d2 := 90 * time.Minute
	got := runtime.Subtract(d1, d2)
	want := d1 - d2
	require.Equal(t, want, got, "Subtract(d1, d2) = %v; want %v", got, want)
}

// TestSubtract_PanicsForOperandCombinationsWithNoDefinedDifference covers:
//
//loop:behavior subtract-subtract-panics-for-operand-combinations-with-no
func TestSubtract_PanicsForOperandCombinationsWithNoDefinedDifference(t *testing.T) {
	tests := []struct {
		name string
		a, b any
	}{
		{"string - string", "foo", "bar"},
		{"bool - bool", true, false},
		{"int - bool", 1, true},
		{"Duration - Time", time.Hour, time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				require.NotNil(t, r, "expected Subtract to panic, got none")
				msg := fmt.Sprint(r)
				assert.Contains(t, msg, "invalid operation", "panic message = %q, want to contain %q", msg, "invalid operation")
			}()
			runtime.Subtract(tt.a, tt.b)
		})
	}
}

// --- Multiply ----------------------------------------------------------------

// TestMultiply_WidensIntegerFamilyOperandsToIntAndFloatOperandsToFloat64 covers:
//
//loop:behavior multiply-multiply-widens-integer-family-operands-to-int-a
func TestMultiply_WidensIntegerFamilyOperandsToIntAndFloatOperandsToFloat64(t *testing.T) {
	tests := []struct {
		name    string
		a, b    any
		wantInt int
		isInt   bool
		wantFlt float64
	}{
		{"int * int", int(2), int(3), 6, true, 0},
		{"uint * int8", uint(2), int8(3), 6, true, 0},
		{"uint8 * uint16", uint8(2), uint16(3), 6, true, 0},
		{"int * float64", int(2), float64(3.5), 0, false, 7.0},
		{"float32 * int", float32(1.5), int(2), 0, false, 3.0},
		{"float64 * float64", float64(1.5), float64(2.5), 0, false, 3.75},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.Multiply(tt.a, tt.b)
			if tt.isInt {
				v, ok := got.(int)
				require.True(t, ok, "Multiply(%v, %v) returned %T; want int", tt.a, tt.b, got)
				require.Equal(t, tt.wantInt, v, "Multiply(%v, %v) = %v; want %v", tt.a, tt.b, v, tt.wantInt)
			} else {
				v, ok := got.(float64)
				require.True(t, ok, "Multiply(%v, %v) returned %T; want float64", tt.a, tt.b, got)
				require.Equal(t, tt.wantFlt, v, "Multiply(%v, %v) = %v; want %v", tt.a, tt.b, v, tt.wantFlt)
			}
		})
	}
}

// TestMultiply_MultipliesEveryPairOfSupportedNumericKinds exercises the full
// cross product of supported numeric kinds so the widening rule is verified
// for every combination, not just a sample.
//
//loop:behavior multiply-multiply-widens-integer-family-operands-to-int-a
func TestMultiply_MultipliesEveryPairOfSupportedNumericKinds(t *testing.T) {
	const av, bv = 3, 4
	for _, ka := range addNumericKinds {
		for _, kb := range addNumericKinds {
			ka, kb := ka, kb
			t.Run(ka.name+"*"+kb.name, func(t *testing.T) {
				a := ka.value(av)
				b := kb.value(bv)
				got := runtime.Multiply(a, b)
				if ka.isFloat || kb.isFloat {
					want := float64(av) * float64(bv)
					v, ok := got.(float64)
					require.True(t, ok, "Multiply(%v, %v) returned %T; want float64", a, b, got)
					require.Equal(t, want, v, "Multiply(%v, %v) = %v; want %v", a, b, v, want)
				} else {
					want := av * bv
					v, ok := got.(int)
					require.True(t, ok, "Multiply(%v, %v) returned %T; want int", a, b, got)
					require.Equal(t, want, v, "Multiply(%v, %v) = %v; want %v", a, b, v, want)
				}
			})
		}
	}
}

// TestMultiply_OfIntegerAndDurationReturnsDurationInEitherOperandOrder covers:
//
//loop:behavior multiply-multiply-of-integer-and-duration-returns-duratio
func TestMultiply_OfIntegerAndDurationReturnsDurationInEitherOperandOrder(t *testing.T) {
	d := 2 * time.Second
	n := 3

	gotIntFirst := runtime.Multiply(n, d)
	gotDurationFirst := runtime.Multiply(d, n)

	want := time.Duration(n) * d
	require.Equal(t, want, gotIntFirst, "Multiply(n, d) = %v; want %v", gotIntFirst, want)
	require.Equal(t, want, gotDurationFirst, "Multiply(d, n) = %v; want %v", gotDurationFirst, want)
}

// TestMultiply_OfEveryIntegerFamilyKindAndDurationReturnsDurationInEitherOrder
// exercises every integer-family kind the behavior names (uint/uint8/…/int64),
// not just int, in both operand orders.
//
//loop:behavior multiply-multiply-of-integer-and-duration-returns-duratio
func TestMultiply_OfEveryIntegerFamilyKindAndDurationReturnsDurationInEitherOrder(t *testing.T) {
	const nv = 3
	d := 2 * time.Second
	want := time.Duration(nv) * d
	for _, k := range addNumericKinds {
		if k.isFloat {
			continue
		}
		k := k
		t.Run(k.name, func(t *testing.T) {
			n := k.value(nv)

			gotIntFirst := runtime.Multiply(n, d)
			require.Equal(t, want, gotIntFirst, "Multiply(%v, %v) = %v; want %v", n, d, gotIntFirst, want)

			gotDurationFirst := runtime.Multiply(d, n)
			require.Equal(t, want, gotDurationFirst, "Multiply(%v, %v) = %v; want %v", d, n, gotDurationFirst, want)
		})
	}
}

// TestMultiply_OfFloatOperandAndDurationReturnsFloat64NotDuration covers:
//
//loop:behavior multiply-multiply-of-a-float-operand-and-a-duration-retur
func TestMultiply_OfFloatOperandAndDurationReturnsFloat64NotDuration(t *testing.T) {
	d := 2 * time.Second
	f := 1.5

	got := runtime.Multiply(f, d)
	v, ok := got.(float64)
	require.True(t, ok, "Multiply(%v, %v) returned %T; want float64", f, d, got)
	want := f * float64(d)
	require.Equal(t, want, v, "Multiply(f, d) = %v; want %v", v, want)
}

// TestMultiply_OfEitherFloatKindAndDurationReturnsFloat64NotDuration covers
// both float32 and float64 operands (the behavior explicitly names both
// kinds as "float32 或 float64"), complementing the float64-only sample
// above.
//
//loop:behavior multiply-multiply-of-a-float-operand-and-a-duration-retur
func TestMultiply_OfEitherFloatKindAndDurationReturnsFloat64NotDuration(t *testing.T) {
	d := 2 * time.Second
	tests := []struct {
		name string
		a    any
		want float64
	}{
		{"float32", float32(1.5), float64(float32(1.5)) * float64(d)},
		{"float64", float64(1.5), 1.5 * float64(d)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.Multiply(tt.a, d)
			v, ok := got.(float64)
			require.True(t, ok, "Multiply(%v, %v) returned %T; want float64", tt.a, d, got)
			require.Equal(t, tt.want, v, "Multiply(%v, %v) = %v; want %v", tt.a, d, v, tt.want)
		})
	}
}

// TestMultiply_OfTwoDurationValuesReturnsTheirProductAsADuration covers:
//
//loop:behavior multiply-multiply-of-two-duration-values-returns-their-pr
func TestMultiply_OfTwoDurationValuesReturnsTheirProductAsADuration(t *testing.T) {
	d1 := 2 * time.Second
	d2 := 3 * time.Second
	got := runtime.Multiply(d1, d2)
	want := time.Duration(d1) * time.Duration(d2)
	require.Equal(t, want, got, "Multiply(d1, d2) = %v; want %v", got, want)
}

// TestMultiply_PanicsForOperandCombinationsWithNoDefinedProduct covers:
//
//loop:behavior multiply-multiply-panics-for-operand-combinations-with-no
func TestMultiply_PanicsForOperandCombinationsWithNoDefinedProduct(t *testing.T) {
	tests := []struct {
		name string
		a, b any
	}{
		{"string * string", "foo", "bar"},
		{"bool * bool", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				require.NotNil(t, r, "expected Multiply to panic, got none")
				msg := fmt.Sprint(r)
				assert.Contains(t, msg, "invalid operation", "panic message = %q, want to contain %q", msg, "invalid operation")
			}()
			runtime.Multiply(tt.a, tt.b)
		})
	}
}

// --- Divide --------------------------------------------------------------------

// TestDivide_AlwaysReturnsAFloat64QuotientEvenForTwoIntegerOperands covers:
//
//loop:behavior divide-divide-always-returns-a-float64-quotient-even-fo
func TestDivide_AlwaysReturnsAFloat64QuotientEvenForTwoIntegerOperands(t *testing.T) {
	tests := []struct {
		name string
		a, b any
		want float64
	}{
		{"int / int, non-exact", int(7), int(2), 3.5},
		{"uint / int8", uint(9), int8(4), 2.25},
		{"float64 / int", float64(7.5), int(2), 3.75},
		{"int / float32", int(9), float32(2), 4.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.Divide(tt.a, tt.b)
			require.Equal(t, tt.want, got, "Divide(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
		})
	}
}

// TestDivide_DividesEveryPairOfSupportedNumericKinds exercises the full cross
// product of supported numeric kinds to confirm the result is always float64.
//
//loop:behavior divide-divide-always-returns-a-float64-quotient-even-fo
func TestDivide_DividesEveryPairOfSupportedNumericKinds(t *testing.T) {
	const av, bv = 9, 4
	for _, ka := range addNumericKinds {
		for _, kb := range addNumericKinds {
			ka, kb := ka, kb
			t.Run(ka.name+"/"+kb.name, func(t *testing.T) {
				a := ka.value(av)
				b := kb.value(bv)
				got := runtime.Divide(a, b)
				want := float64(av) / float64(bv)
				require.Equal(t, want, got, "Divide(%v, %v) = %v; want %v", a, b, got, want)
			})
		}
	}
}

// TestDivide_PanicsForOperandCombinationsWithNoDefinedQuotient covers:
//
//loop:behavior divide-divide-panics-for-operand-combinations-with-no-d
func TestDivide_PanicsForOperandCombinationsWithNoDefinedQuotient(t *testing.T) {
	tests := []struct {
		name string
		a, b any
	}{
		{"string / string", "foo", "bar"},
		{"bool / bool", true, false},
		{"Time / Time", time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC), time.Date(2021, time.January, 2, 0, 0, 0, 0, time.UTC)},
		{"Duration / Duration", time.Hour, time.Minute},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				require.NotNil(t, r, "expected Divide to panic, got none")
				msg := fmt.Sprint(r)
				assert.Contains(t, msg, "invalid operation", "panic message = %q, want to contain %q", msg, "invalid operation")
			}()
			runtime.Divide(tt.a, tt.b)
		})
	}
}

// --- Modulo --------------------------------------------------------------------

// TestModulo_ReturnsAnIntRemainderForTwoIntegerFamilyOperands covers:
//
//loop:behavior modulo-modulo-returns-an-int-remainder-for-two-integer
func TestModulo_ReturnsAnIntRemainderForTwoIntegerFamilyOperands(t *testing.T) {
	tests := []struct {
		name string
		a, b any
		want int
	}{
		{"int % int", int(7), int(3), 1},
		{"uint % int8", uint(9), int8(4), 1},
		{"uint8 % uint16", uint8(10), uint16(3), 1},
		{"int64 % int32", int64(20), int32(6), 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.Modulo(tt.a, tt.b)
			require.Equal(t, tt.want, got, "Modulo(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
		})
	}
}

// TestModulo_ModulosEveryPairOfSupportedIntegerFamilyKinds exercises the full
// cross product of the integer-family kinds Modulo accepts (unlike
// Add/Subtract/Multiply/Divide, floats are excluded from this set).
//
//loop:behavior modulo-modulo-returns-an-int-remainder-for-two-integer
func TestModulo_ModulosEveryPairOfSupportedIntegerFamilyKinds(t *testing.T) {
	const av, bv = 10, 3
	for _, ka := range addNumericKinds {
		if ka.isFloat {
			continue
		}
		for _, kb := range addNumericKinds {
			if kb.isFloat {
				continue
			}
			ka, kb := ka, kb
			t.Run(ka.name+"%"+kb.name, func(t *testing.T) {
				a := ka.value(av)
				b := kb.value(bv)
				got := runtime.Modulo(a, b)
				want := av % bv
				require.Equal(t, want, got, "Modulo(%v, %v) = %v; want %v", a, b, got, want)
			})
		}
	}
}

// TestModulo_PanicsOnFloatOperandsUnlikeAddSubtractMultiplyDivide covers:
//
//loop:behavior modulo-modulo-panics-on-float-operands-unlike-add-subtr
func TestModulo_PanicsOnFloatOperandsUnlikeAddSubtractMultiplyDivide(t *testing.T) {
	tests := []struct {
		name string
		a, b any
	}{
		{"a is float64", float64(7), int(3)},
		{"b is float32", int(7), float32(3)},
		{"both float", float64(7), float64(3)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				require.NotNil(t, r, "expected Modulo to panic, got none")
				msg := fmt.Sprint(r)
				assert.Contains(t, msg, "invalid operation", "panic message = %q, want to contain %q", msg, "invalid operation")
			}()
			runtime.Modulo(tt.a, tt.b)
		})
	}
}

// TestModulo_PanicsForNonNumericOperands covers:
//
//loop:behavior modulo-modulo-panics-for-non-numeric-operands
func TestModulo_PanicsForNonNumericOperands(t *testing.T) {
	tests := []struct {
		name string
		a, b any
	}{
		{"string % string", "foo", "bar"},
		{"bool % bool", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				require.NotNil(t, r, "expected Modulo to panic, got none")
				msg := fmt.Sprint(r)
				assert.Contains(t, msg, "invalid operation", "panic message = %q, want to contain %q", msg, "invalid operation")
			}()
			runtime.Modulo(tt.a, tt.b)
		})
	}
}

// --- LessOrEqual -----------------------------------------------------------

// TestLessOrEqual_ComparesNumericOperandsWidenedToIntOrFloat64 covers:
//
//loop:behavior lessorequal-lessorequal-compares-numeric-operands-widened-to
func TestLessOrEqual_ComparesNumericOperandsWidenedToIntOrFloat64(t *testing.T) {
	tests := []struct {
		name string
		a, b any
		want bool
	}{
		{"int < int", int(2), int(3), true},
		{"int == int", int(3), int(3), true},
		{"int > int", int(4), int(3), false},
		{"uint < int8", uint(2), int8(3), true},
		{"float64 == int", float64(3), int(3), true},
		{"float32 > int", float32(4.5), int(4), false},
		{"int < float64", int(3), float64(3.5), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.LessOrEqual(tt.a, tt.b)
			require.Equal(t, tt.want, got, "LessOrEqual(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
		})
	}
}

// TestLessOrEqual_ComparesEveryPairOfSupportedNumericKinds exercises the full
// cross product of supported numeric kinds (mirroring the *EveryPair tests
// for Add/Subtract/Multiply/Divide/Modulo above) so the widening rule is
// verified for every combination, not just a handful of representative
// types, and for both the "less" and "equal" branches of "<=".
//
//loop:behavior lessorequal-lessorequal-compares-numeric-operands-widened-to
func TestLessOrEqual_ComparesEveryPairOfSupportedNumericKinds(t *testing.T) {
	for _, ka := range addNumericKinds {
		for _, kb := range addNumericKinds {
			ka, kb := ka, kb
			t.Run(ka.name+"<="+kb.name, func(t *testing.T) {
				// a < b: 3 <= 4 must be true.
				aLess := ka.value(3)
				bLess := kb.value(4)
				gotLess := runtime.LessOrEqual(aLess, bLess)
				require.True(t, gotLess, "LessOrEqual(%v, %v) = false; want true", aLess, bLess)

				// a == b: 4 <= 4 must be true (inclusive of equality).
				aEq := ka.value(4)
				bEq := kb.value(4)
				gotEq := runtime.LessOrEqual(aEq, bEq)
				require.True(t, gotEq, "LessOrEqual(%v, %v) = false; want true", aEq, bEq)

				// a > b: 4 <= 3 must be false.
				aMore := ka.value(4)
				bMore := kb.value(3)
				gotMore := runtime.LessOrEqual(aMore, bMore)
				require.False(t, gotMore, "LessOrEqual(%v, %v) = true; want false", aMore, bMore)
			})
		}
	}
}

// TestLessOrEqual_ComparesTwoStringOperandsLexicographicallyInclusiveOfEquality covers:
//
//loop:behavior lessorequal-lessorequal-compares-two-string-operands-lexicog
func TestLessOrEqual_ComparesTwoStringOperandsLexicographicallyInclusiveOfEquality(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{"equal", "abc", "abc", true},
		{"a < b", "abc", "abd", true},
		{"a > b", "abd", "abc", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.LessOrEqual(tt.a, tt.b)
			require.Equal(t, tt.want, got, "LessOrEqual(%q, %q) = %v; want %v", tt.a, tt.b, got, tt.want)
		})
	}
}

// TestLessOrEqual_ComparesTwoTimeOperandsChronologicallyInclusiveOfEquality covers:
//
//loop:behavior lessorequal-lessorequal-compares-two-time-operands-chronolog
func TestLessOrEqual_ComparesTwoTimeOperandsChronologicallyInclusiveOfEquality(t *testing.T) {
	earlier := time.Date(2021, time.January, 1, 10, 0, 0, 0, time.UTC)
	later := time.Date(2021, time.January, 1, 12, 0, 0, 0, time.UTC)
	sameInstant := time.Date(2021, time.January, 1, 10, 0, 0, 0, time.UTC)

	require.True(t, runtime.LessOrEqual(earlier, later), "LessOrEqual(earlier, later) = false; want true")
	require.False(t, runtime.LessOrEqual(later, earlier), "LessOrEqual(later, earlier) = true; want false")
	require.True(t, runtime.LessOrEqual(earlier, sameInstant), "LessOrEqual(earlier, sameInstant) = false; want true")
}

// TestLessOrEqual_ComparesTwoDurationOperandsNumericallyInclusiveOfEquality covers:
//
//loop:behavior lessorequal-lessorequal-compares-two-duration-operands-numer
func TestLessOrEqual_ComparesTwoDurationOperandsNumericallyInclusiveOfEquality(t *testing.T) {
	tests := []struct {
		name string
		a, b time.Duration
		want bool
	}{
		{"equal", time.Hour, time.Hour, true},
		{"a < b", time.Minute, time.Hour, true},
		{"a > b", time.Hour, time.Minute, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtime.LessOrEqual(tt.a, tt.b)
			require.Equal(t, tt.want, got, "LessOrEqual(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
		})
	}
}

// TestLessOrEqual_PanicsForOperandCombinationsWithNoDefinedOrder covers:
//
//loop:behavior lessorequal-lessorequal-panics-for-operand-combinations-with
func TestLessOrEqual_PanicsForOperandCombinationsWithNoDefinedOrder(t *testing.T) {
	tests := []struct {
		name string
		a, b any
	}{
		{"bool <= bool", true, false},
		{"int <= bool", 1, true},
		{"string <= int", "foo", 1},
		{"Time <= Duration", time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC), time.Hour},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				require.NotNil(t, r, "expected LessOrEqual to panic, got none")
				msg := fmt.Sprint(r)
				assert.Contains(t, msg, "invalid operation", "panic message = %q, want to contain %q", msg, "invalid operation")
			}()
			runtime.LessOrEqual(tt.a, tt.b)
		})
	}
}
