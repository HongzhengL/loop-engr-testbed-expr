package value_test

import (
	"testing"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/internal/testify/require"
	"github.com/expr-lang/expr/patcher/value"
)

// plainNoValuer implements none of the value package's Valuer interfaces.
type plainNoValuer struct {
	Int int
}

func TestValueGetter_ValueImplementingNoSupportedValuerInterfaceIsUsedUnmodified(t *testing.T) {
	// loop:behavior valuegetter-value-implementing-no-supported-valuer-interface
	env := map[string]any{"V": plainNoValuer{Int: 5}}

	program, err := expr.Compile("V", expr.Env(env), value.ValueGetter)
	require.NoError(t, err)
	require.Equal(t, "V", program.Node().String(), "want no conversion call inserted for a non-valuer type")

	out, err := expr.Run(program, env)
	require.NoError(t, err)
	require.Equal(t, plainNoValuer{Int: 5}, out)
}

// anyOnlyValuer implements only AnyValuer, and returns a nil value from it.
type anyOnlyValuer struct{}

func (anyOnlyValuer) AsAny() any { return nil }

func TestValueGetter_AnyValuerMayReturnANilValueWithoutACompileTimeTypeCheck(t *testing.T) {
	// loop:behavior valuegetter-anyvaluer-may-return-a-nil-value-without-a-compi
	env := map[string]any{"V": anyOnlyValuer{}}

	program, err := expr.Compile("V", expr.Env(env), value.ValueGetter)
	require.NoError(t, err, "compilation must succeed even though AsAny() returns nil")

	out, err := expr.Run(program, env)
	require.NoError(t, err)
	require.Nil(t, out)
}

// dualValuer implements both AnyValuer and IntValuer, per the doc comment's
// dual-implementation pattern. Both accessors are made to agree on the same
// externally observable value (7) so the assertion does not depend on which
// of the two the package chooses at run time -- that choice is unspecified.
type dualValuer struct{}

func (dualValuer) AsAny() any { return 7 }
func (dualValuer) AsInt() int { return 7 }

func TestValueGetter_TypeImplementingBothAnyValuerAndATypedValuerStillCompiles(t *testing.T) {
	// loop:behavior valuegetter-type-implementing-both-anyvaluer-and-a-typed-val
	env := map[string]any{"V": dualValuer{}}

	program, err := expr.Compile("V", expr.Env(env), value.ValueGetter)
	require.NoError(t, err)

	out, err := expr.Run(program, env)
	require.NoError(t, err)
	require.Equal(t, 7, out)
}

type boolValuer struct{ B bool }

func (v boolValuer) AsBool() bool { return v.B }

func TestValueGetter_BoolValuerParticipatesInABooleanExpression(t *testing.T) {
	// loop:behavior valuegetter-boolvaluer-participates-in-a-boolean-expression
	env := map[string]any{"V": boolValuer{B: true}}

	program, err := expr.Compile("!V", expr.Env(env), value.ValueGetter)
	require.NoError(t, err)

	out, err := expr.Run(program, env)
	require.NoError(t, err)
	require.Equal(t, false, out)
}

type timeValuer struct{ T time.Time }

func (v timeValuer) AsTime() time.Time { return v.T }

func TestValueGetter_TimeValuerParticipatesInATimeExpression(t *testing.T) {
	// loop:behavior valuegetter-timevaluer-participates-in-a-time-expression
	env := map[string]any{
		"V":  timeValuer{T: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)},
		"T2": time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	program, err := expr.Compile("V > T2", expr.Env(env), value.ValueGetter)
	require.NoError(t, err)

	out, err := expr.Run(program, env)
	require.NoError(t, err)
	require.Equal(t, true, out)
}

type durationValuer struct{ D time.Duration }

func (v durationValuer) AsDuration() time.Duration { return v.D }

func TestValueGetter_DurationValuerParticipatesInADurationExpression(t *testing.T) {
	// loop:behavior valuegetter-durationvaluer-participates-in-a-duration-expres
	env := map[string]any{
		"V":  durationValuer{D: 2 * time.Second},
		"D2": time.Second,
	}

	program, err := expr.Compile("V > D2", expr.Env(env), value.ValueGetter)
	require.NoError(t, err)

	out, err := expr.Run(program, env)
	require.NoError(t, err)
	require.Equal(t, true, out)
}
