package nature_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/expr-lang/expr/checker/nature"
)

// anyType is the reflect.Type for the empty interface (interface{} / any).
var anyType = reflect.TypeOf((*any)(nil)).Elem()

// MyBytes is a named type whose underlying type is []byte, distinct from the
// exact []byte type, used to test that IsByteSlice() is identity-based.
type MyBytes []byte

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

//loop:behavior nature-isbool-isbool-true-for-bool-kind
func TestNature_IsBool_TrueForBoolKind(t *testing.T) {
	var c *nature.Cache
	n := c.FromType(reflect.TypeOf(true))

	if got := n.IsBool(); !got {
		t.Errorf("FromType(bool).IsBool() = %v, want true", got)
	}
}

//loop:behavior nature-isstring-isstring-true-for-string-kind
func TestNature_IsString_TrueForStringKind(t *testing.T) {
	var c *nature.Cache
	n := c.FromType(reflect.TypeOf(""))

	if got := n.IsString(); !got {
		t.Errorf("FromType(string).IsString() = %v, want true", got)
	}
}

//loop:behavior nature-isarray-isarray-true-for-slice-or-array-kind
func TestNature_IsArray_TrueForSliceOrArrayKind(t *testing.T) {
	var c *nature.Cache
	cases := map[string]reflect.Type{
		"slice": reflect.TypeOf([]string{}),
		"array": reflect.TypeOf([2]string{}),
	}
	for name, tp := range cases {
		t.Run(name, func(t *testing.T) {
			n := c.FromType(tp)
			if got := n.IsArray(); !got {
				t.Errorf("FromType(%v).IsArray() = %v, want true", tp, got)
			}
		})
	}
}

//loop:behavior nature-isarray-isarray-false-for-non-array-non-slice-kind
func TestNature_IsArray_FalseForNonArrayNonSliceKind(t *testing.T) {
	var c *nature.Cache
	n := c.FromType(reflect.TypeOf(0))

	if got := n.IsArray(); got {
		t.Errorf("FromType(int).IsArray() = %v, want false", got)
	}
}

//loop:behavior nature-ismap-ismap-true-for-map-kind
func TestNature_IsMap_TrueForMapKind(t *testing.T) {
	var c *nature.Cache
	n := c.FromType(reflect.TypeOf(map[string]int{}))

	if got := n.IsMap(); !got {
		t.Errorf("FromType(map[string]int).IsMap() = %v, want true", got)
	}
}

//loop:behavior nature-isstruct-isstruct-true-for-struct-kind
func TestNature_IsStruct_TrueForStructKind(t *testing.T) {
	var c *nature.Cache
	n := c.FromType(reflect.TypeOf(struct{}{}))

	if got := n.IsStruct(); !got {
		t.Errorf("FromType(struct{}{}).IsStruct() = %v, want true", got)
	}
}

//loop:behavior nature-isfunc-isfunc-true-for-func-kind
func TestNature_IsFunc_TrueForFuncKind(t *testing.T) {
	var c *nature.Cache
	n := c.FromType(reflect.TypeOf(func() {}))

	if got := n.IsFunc(); !got {
		t.Errorf("FromType(func()).IsFunc() = %v, want true", got)
	}
}

//loop:behavior nature-ispointer-ispointer-true-for-pointer-kind
func TestNature_IsPointer_TrueForPointerKind(t *testing.T) {
	var c *nature.Cache
	n := c.FromType(reflect.TypeOf(new(int)))

	if got := n.IsPointer(); !got {
		t.Errorf("FromType(*int).IsPointer() = %v, want true", got)
	}
}

//loop:behavior nature-isbyteslice-isbyteslice-true-only-for-the-exact-byte-type
func TestNature_IsByteSlice_TrueForExactByteSliceType(t *testing.T) {
	var c *nature.Cache
	n := c.FromType(reflect.TypeOf([]byte{}))

	if got := n.IsByteSlice(); !got {
		t.Errorf("FromType([]byte{}).IsByteSlice() = %v, want true", got)
	}
}

//loop:behavior nature-isbyteslice-isbyteslice-false-for-a-distinct-named-byte-slic
func TestNature_IsByteSlice_FalseForDistinctNamedByteSliceType(t *testing.T) {
	var c *nature.Cache
	cases := map[string]reflect.Type{
		"named byte slice type (MyBytes)": reflect.TypeOf(MyBytes{}),
		"unrelated slice type ([]string)": reflect.TypeOf([]string{}),
	}
	for name, tp := range cases {
		t.Run(name, func(t *testing.T) {
			n := c.FromType(tp)
			if got := n.IsByteSlice(); got {
				t.Errorf("FromType(%v).IsByteSlice() = %v, want false", tp, got)
			}
		})
	}
}

//loop:behavior nature-istime-istime-true-only-for-the-exact-time-time-type
func TestNature_IsTime_TrueOnlyForExactTimeType(t *testing.T) {
	var c *nature.Cache

	timeNature := c.FromType(reflect.TypeOf(time.Time{}))
	if got := timeNature.IsTime(); !got {
		t.Errorf("FromType(time.Time{}).IsTime() = %v, want true", got)
	}

	intNature := c.FromType(reflect.TypeOf(0))
	if got := intNature.IsTime(); got {
		t.Errorf("FromType(int).IsTime() = %v, want false", got)
	}
}

//loop:behavior nature-isduration-isduration-true-only-for-the-exact-time-duration
func TestNature_IsDuration_TrueOnlyForExactDurationType(t *testing.T) {
	var c *nature.Cache

	durationNature := c.FromType(reflect.TypeOf(time.Duration(0)))
	if got := durationNature.IsDuration(); !got {
		t.Errorf("FromType(time.Duration(0)).IsDuration() = %v, want true", got)
	}

	int64Nature := c.FromType(reflect.TypeOf(int64(0)))
	if got := int64Nature.IsDuration(); got {
		t.Errorf("FromType(int64).IsDuration() = %v, want false", got)
	}
}

//loop:behavior nature-isnumber-isnumber-true-when-isinteger-is-set
func TestNature_IsNumber_TrueWhenIsIntegerIsSet(t *testing.T) {
	var c *nature.Cache
	n := c.FromType(reflect.TypeOf(int(0)))

	if !n.IsInteger || n.IsFloat {
		t.Fatalf("precondition failed: FromType(int) IsInteger=%v IsFloat=%v, want IsInteger=true IsFloat=false", n.IsInteger, n.IsFloat)
	}
	if got := n.IsNumber(); !got {
		t.Errorf("IsNumber() with IsInteger=true = %v, want true", got)
	}
}

//loop:behavior nature-isnumber-isnumber-true-when-isfloat-is-set
func TestNature_IsNumber_TrueWhenIsFloatIsSet(t *testing.T) {
	var c *nature.Cache
	n := c.FromType(reflect.TypeOf(float64(0)))

	if !n.IsFloat || n.IsInteger {
		t.Fatalf("precondition failed: FromType(float64) IsFloat=%v IsInteger=%v, want IsFloat=true IsInteger=false", n.IsFloat, n.IsInteger)
	}
	if got := n.IsNumber(); !got {
		t.Errorf("IsNumber() with IsFloat=true = %v, want true", got)
	}
}

//loop:behavior nature-isnumber-isnumber-false-when-neither-isinteger-nor-isfloa
func TestNature_IsNumber_FalseWhenNeitherIsIntegerNorIsFloatIsSet(t *testing.T) {
	var c *nature.Cache
	n := c.FromType(reflect.TypeOf(""))

	if n.IsInteger || n.IsFloat {
		t.Fatalf("precondition failed: FromType(string) IsInteger=%v IsFloat=%v, want both false", n.IsInteger, n.IsFloat)
	}
	if got := n.IsNumber(); got {
		t.Errorf("IsNumber() with IsInteger=false IsFloat=false = %v, want false", got)
	}
}

//loop:behavior nature-string-string-returns-the-underlying-reflect-type-strin
func TestNature_String_ReturnsUnderlyingReflectTypeStringWhenTypeIsSet(t *testing.T) {
	var c *nature.Cache
	tp := reflect.TypeOf(0)
	n := c.FromType(tp)

	got := n.String()
	want := tp.String()
	if got != want {
		t.Errorf("Nature.String() = %q, want %q", got, want)
	}
}

//loop:behavior nature-string-string-returns-the-literal-unknown-when-type-is
func TestNature_String_ReturnsUnknownLiteralWhenTypeIsNil(t *testing.T) {
	n := nature.Nature{}

	got := n.String()
	want := "unknown"
	if got != want {
		t.Errorf("Nature{}.String() = %q, want %q", got, want)
	}
}

//loop:behavior natureof-package-natureof-nil-input-returns-nil-nature
func TestNatureOf_PackageLevel_NilInputReturnsNilNature(t *testing.T) {
	got := nature.NatureOf(nil)

	if !got.Nil {
		t.Errorf("nature.NatureOf(nil).Nil = %v, want true", got.Nil)
	}
	if got.Type != nil {
		t.Errorf("nature.NatureOf(nil).Type = %v, want nil", got.Type)
	}
}

//loop:behavior natureof-package-natureof-non-nil-input-matches-fromtype
func TestNatureOf_PackageLevel_NonNilInputMatchesFromTypeOfValueType(t *testing.T) {
	i := 3

	got := nature.NatureOf(i)
	want := nature.FromType(reflect.TypeOf(i))

	if got != want {
		t.Errorf("nature.NatureOf(%v) = %+v, want %+v (nature.FromType(reflect.TypeOf(%v)))", i, got, want, i)
	}
}

//loop:behavior fromtype-package-fromtype-nil-reflect-type-returns-unknow
func TestFromType_PackageLevel_NilReflectTypeReturnsUnknownZeroValueNature(t *testing.T) {
	got := nature.FromType(nil)

	if got.Type != nil {
		t.Errorf("nature.FromType(nil).Type = %v, want nil", got.Type)
	}
	if got.TypeData != nil {
		t.Errorf("nature.FromType(nil).TypeData = %v, want nil", got.TypeData)
	}
}
