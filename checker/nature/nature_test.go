package nature_test

import (
	"reflect"
	"testing"

	"github.com/expr-lang/expr/checker/nature"
)

// anyType is the reflect.Type for the empty interface (interface{} / any).
var anyType = reflect.TypeOf((*any)(nil)).Elem()

//loop:behavior cache-fromtype-fromtype-nil-reflect-type-returns-unknown-zero-v
func TestCache_FromType_NilTypeReturnsUnknownNature(t *testing.T) {
	caches := map[string]*nature.Cache{
		"nil cache":     nil,
		"non-nil cache": {},
	}
	for name, c := range caches {
		t.Run(name, func(t *testing.T) {
			got := c.FromType(nil)
			if got.Type != nil {
				t.Errorf("FromType(nil).Type = %v, want nil", got.Type)
			}
			if got.TypeData != nil {
				t.Errorf("FromType(nil).TypeData = %v, want nil", got.TypeData)
			}
		})
	}
}

//loop:behavior cache-fromtype-fromtype-non-nil-type-sets-type-and-kind-fields
func TestCache_FromType_NonNilTypeSetsTypeAndKind(t *testing.T) {
	var c *nature.Cache
	tp := reflect.TypeOf(0)

	got := c.FromType(tp)

	if got.Type != tp {
		t.Errorf("FromType(%v).Type = %v, want %v", tp, got.Type, tp)
	}
	if got.Kind != reflect.Int {
		t.Errorf("FromType(%v).Kind = %v, want %v", tp, got.Kind, reflect.Int)
	}
}

//loop:behavior cache-natureof-natureof-nil-interface-input-returns-nil-nature
func TestCache_NatureOf_NilInputReturnsNilNature(t *testing.T) {
	caches := map[string]*nature.Cache{
		"nil cache":     nil,
		"non-nil cache": {},
	}
	for name, c := range caches {
		t.Run(name, func(t *testing.T) {
			got := c.NatureOf(nil)
			if !got.Nil {
				t.Errorf("NatureOf(nil).Nil = false, want true")
			}
			if got.Type != nil {
				t.Errorf("NatureOf(nil).Type = %v, want nil", got.Type)
			}
		})
	}
}

//loop:behavior cache-natureof-natureof-non-nil-input-delegates-to-fromtype-of
func TestCache_NatureOf_NonNilInputDelegatesToFromType(t *testing.T) {
	var c *nature.Cache
	i := 3

	got := c.NatureOf(i)
	want := c.FromType(reflect.TypeOf(i))

	if got != want {
		t.Errorf("NatureOf(%v) = %+v, want %+v (FromType(reflect.TypeOf(%v)))", i, got, want, i)
	}
}

//loop:behavior nature-isunknown-isunknown-reports-true-for-zero-value-nature
func TestNature_IsUnknown_ZeroValueReportsTrue(t *testing.T) {
	var c *nature.Cache
	n := nature.Nature{}

	if got := n.IsUnknown(c); !got {
		t.Errorf("Nature{}.IsUnknown(c) = %v, want true", got)
	}
}

//loop:behavior nature-isunknown-isunknown-reports-false-for-explicit-nil-value-n
func TestNature_IsUnknown_ExplicitNilReportsFalse(t *testing.T) {
	var c *nature.Cache
	n := nature.Nature{Nil: true}

	if got := n.IsUnknown(c); got {
		t.Errorf("Nature{Nil: true}.IsUnknown(c) = %v, want false", got)
	}
}

//loop:behavior nature-isany-isany-reports-true-for-empty-interface-type
func TestNature_IsAny_EmptyInterfaceReportsTrue(t *testing.T) {
	var c *nature.Cache
	n := c.FromType(anyType)

	if got := n.IsAny(c); !got {
		t.Errorf("FromType(any).IsAny(c) = %v, want true", got)
	}
}

//loop:behavior nature-isany-isany-reports-false-for-concrete-non-interface-t
func TestNature_IsAny_ConcreteTypeReportsFalse(t *testing.T) {
	var c *nature.Cache
	n := c.FromType(reflect.TypeOf(0))

	if got := n.IsAny(c); got {
		t.Errorf("FromType(int).IsAny(c) = %v, want false", got)
	}
}

//loop:behavior nature-assignableto-assignableto-nil-nature-is-assignable-to-referen
func TestNature_AssignableTo_NilNatureAssignableToReferenceKinds(t *testing.T) {
	n := nature.Nature{Nil: true}
	kinds := []reflect.Kind{
		reflect.Pointer,
		reflect.Interface,
		reflect.Chan,
		reflect.Func,
		reflect.Map,
		reflect.Slice,
	}
	for _, k := range kinds {
		t.Run(k.String(), func(t *testing.T) {
			target := nature.Nature{Kind: k}
			if got := n.AssignableTo(target); !got {
				t.Errorf("Nature{Nil: true}.AssignableTo(Nature{Kind: %v}) = %v, want true", k, got)
			}
		})
	}
}

//loop:behavior nature-assignableto-assignableto-nil-nature-is-not-assignable-to-val
func TestNature_AssignableTo_NilNatureNotAssignableToValueKinds(t *testing.T) {
	n := nature.Nature{Nil: true}
	kinds := []reflect.Kind{
		reflect.Struct,
		reflect.Int,
	}
	for _, k := range kinds {
		t.Run(k.String(), func(t *testing.T) {
			target := nature.Nature{Kind: k}
			if got := n.AssignableTo(target); got {
				t.Errorf("Nature{Nil: true}.AssignableTo(Nature{Kind: %v}) = %v, want false", k, got)
			}
		})
	}
}

//loop:behavior nature-assignableto-assignableto-unknown-nature-is-never-assignable
func TestNature_AssignableTo_UnknownNatureNeverAssignable(t *testing.T) {
	var c *nature.Cache
	n := nature.Nature{}
	target := c.FromType(reflect.TypeOf(0))

	if got := n.AssignableTo(target); got {
		t.Errorf("Nature{}.AssignableTo(FromType(int)) = %v, want false", got)
	}
}

//loop:behavior nature-assignableto-assignableto-concrete-type-is-assignable-to-empt
func TestNature_AssignableTo_ConcreteTypeAssignableToEmptyInterface(t *testing.T) {
	var c *nature.Cache
	n := c.FromType(reflect.TypeOf(0))
	target := c.FromType(anyType)

	if got := n.AssignableTo(target); !got {
		t.Errorf("FromType(int).AssignableTo(FromType(any)) = %v, want true", got)
	}
}

//loop:behavior nature-assignableto-assignableto-different-kind-non-interface-target
func TestNature_AssignableTo_DifferentKindNonInterfaceTargetFalse(t *testing.T) {
	var c *nature.Cache
	n := c.FromType(reflect.TypeOf(0))
	target := c.FromType(reflect.TypeOf(""))

	if got := n.AssignableTo(target); got {
		t.Errorf("FromType(int).AssignableTo(FromType(string)) = %v, want false", got)
	}
}

//loop:behavior nature-comparableto-comparableto-unknown-or-nil-operand-always-repor
func TestNature_ComparableTo_UnknownOrNilOperandAlwaysTrue(t *testing.T) {
	var c *nature.Cache
	concrete := c.FromType(reflect.TypeOf(0))

	cases := map[string]struct {
		lhs, rhs nature.Nature
	}{
		"lhs unknown":      {lhs: nature.Nature{}, rhs: concrete},
		"rhs unknown":      {lhs: concrete, rhs: nature.Nature{}},
		"lhs explicit nil": {lhs: nature.Nature{Nil: true}, rhs: concrete},
		"rhs explicit nil": {lhs: concrete, rhs: nature.Nature{Nil: true}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := tc.lhs.ComparableTo(c, tc.rhs); !got {
				t.Errorf("ComparableTo = %v, want true", got)
			}
		})
	}
}

//loop:behavior nature-comparableto-comparableto-cross-numeric-type-operands-report
func TestNature_ComparableTo_CrossNumericTypesReportsTrue(t *testing.T) {
	var c *nature.Cache
	n := c.FromType(reflect.TypeOf(int(0)))
	rhs := c.FromType(reflect.TypeOf(float64(0)))

	if got := n.ComparableTo(c, rhs); !got {
		t.Errorf("int.ComparableTo(float64) = %v, want true", got)
	}
}

//loop:behavior nature-comparableto-comparableto-unrelated-non-numeric-non-assignabl
func TestNature_ComparableTo_UnrelatedNonNumericNonAssignableReportsFalse(t *testing.T) {
	var c *nature.Cache
	n := c.FromType(reflect.TypeOf(""))
	rhs := c.FromType(reflect.TypeOf(struct{}{}))

	if got := n.ComparableTo(c, rhs); got {
		t.Errorf("string.ComparableTo(struct{}) = %v, want false", got)
	}
}

//loop:behavior nature-elem-elem-pointer-kind-returns-nature-of-pointed-to-t
func TestNature_Elem_PointerKindReturnsPointedToType(t *testing.T) {
	var c *nature.Cache
	n := c.FromType(reflect.TypeOf(new(int)))

	got := n.Elem(c)

	want := reflect.TypeOf(int(0))
	if got.Type != want {
		t.Errorf("Elem().Type = %v, want %v", got.Type, want)
	}
}

//loop:behavior nature-elem-elem-map-kind-with-defaultmapvalue-overrides-dec
func TestNature_Elem_MapKindWithDefaultMapValueOverridesDeclaredType(t *testing.T) {
	var c *nature.Cache
	custom := nature.Nature{Type: reflect.TypeOf("")}
	n := c.FromType(reflect.TypeOf(map[string]int{}))
	n.TypeData = &nature.TypeData{DefaultMapValue: &custom}

	got := n.Elem(c)

	if got != custom {
		t.Errorf("Elem() = %+v, want DefaultMapValue %+v", got, custom)
	}
	if got.Type == reflect.TypeOf(int(0)) {
		t.Errorf("Elem().Type = %v, should not be the map's declared value type (int)", got.Type)
	}
}

//loop:behavior nature-elem-elem-map-kind-without-defaultmapvalue-falls-back
func TestNature_Elem_MapKindWithoutDefaultMapValueFallsBackToDeclaredType(t *testing.T) {
	var c *nature.Cache
	n := c.FromType(reflect.TypeOf(map[string]int{}))

	got := n.Elem(c)

	want := reflect.TypeOf(int(0))
	if got.Type != want {
		t.Errorf("Elem().Type = %v, want %v", got.Type, want)
	}
}

//loop:behavior nature-elem-elem-slice-or-array-kind-with-ref-overrides-decl
func TestNature_Elem_SliceOrArrayKindWithRefOverridesDeclaredElementType(t *testing.T) {
	var c *nature.Cache
	custom := nature.Nature{Type: reflect.TypeOf(0)}

	cases := map[string]reflect.Type{
		"slice": reflect.TypeOf([]string{}),
		"array": reflect.TypeOf([2]string{}),
	}
	for name, tp := range cases {
		t.Run(name, func(t *testing.T) {
			n := c.FromType(tp)
			n.Ref = &custom

			got := n.Elem(c)

			if got != custom {
				t.Errorf("Elem() = %+v, want Ref %+v", got, custom)
			}
			if got.Type == reflect.TypeOf("") {
				t.Errorf("Elem().Type = %v, should not be the declared element type (string)", got.Type)
			}
		})
	}
}
