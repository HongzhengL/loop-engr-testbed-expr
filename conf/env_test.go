package conf_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/expr-lang/expr/checker/nature"
	"github.com/expr-lang/expr/conf"
	"github.com/expr-lang/expr/types"
)

// TestEnvWithCache_NilEnv verifies that a nil env is treated as an empty
// strict map.
//
//loop:behavior envwithcache-envwithcache-treats-nil-env-as-empty-strict-map
func TestEnvWithCache_NilEnv(t *testing.T) {
	t.Run("zero-valued cache", func(t *testing.T) {
		c := &nature.Cache{}
		got := conf.EnvWithCache(c, nil)

		if got.Kind != reflect.Map {
			t.Errorf("Kind = %v, want %v", got.Kind, reflect.Map)
		}
		if !got.Strict {
			t.Errorf("Strict = %v, want true", got.Strict)
		}
	})

	t.Run("new cache", func(t *testing.T) {
		c := new(nature.Cache)
		got := conf.EnvWithCache(c, nil)

		if got.Kind != reflect.Map {
			t.Errorf("Kind = %v, want %v", got.Kind, reflect.Map)
		}
		if !got.Strict {
			t.Errorf("Strict = %v, want true", got.Strict)
		}
	})
}

// TestEnvWithCache_StructEnv verifies that a non-pointer struct env is
// marked as Strict, and that Type/Kind reflect the struct's own type.
//
//loop:behavior envwithcache-envwithcache-marks-struct-env-as-strict
func TestEnvWithCache_StructEnv(t *testing.T) {
	env := struct{ X int }{}
	c := &nature.Cache{}

	got := conf.EnvWithCache(c, env)

	if got.Kind != reflect.Struct {
		t.Errorf("Kind = %v, want %v", got.Kind, reflect.Struct)
	}
	wantType := reflect.TypeOf(env)
	if got.Type != wantType {
		t.Errorf("Type = %v, want %v", got.Type, wantType)
	}
	if !got.Strict {
		t.Errorf("Strict = %v, want true", got.Strict)
	}
}

// TestEnvWithCache_MapEnv_FieldsPerKey verifies that a native map env
// produces one Fields entry per map key.
//
//loop:behavior envwithcache-envwithcache-builds-one-fields-entry-per-map-key
func TestEnvWithCache_MapEnv_FieldsPerKey(t *testing.T) {
	env := map[string]int{"a": 1, "b": 2}
	c := &nature.Cache{}

	got := conf.EnvWithCache(c, env)

	if got.Kind != reflect.Map {
		t.Fatalf("Kind = %v, want %v", got.Kind, reflect.Map)
	}
	if !got.Strict {
		t.Errorf("Strict = %v, want true", got.Strict)
	}
	if len(got.Fields) != len(env) {
		t.Fatalf("len(Fields) = %d, want %d", len(got.Fields), len(env))
	}
	for key := range env {
		if _, ok := got.Fields[key]; !ok {
			t.Errorf("Fields[%q] missing, want an entry for every map key", key)
		}
	}
}

// TestEnvWithCache_MapEnv_NilInterfaceValue verifies that a nil interface
// value stored in a map env is recorded as Nature.Nil == true.
//
//loop:behavior envwithcache-envwithcache-records-nil-map-value-as-nature-nil
func TestEnvWithCache_MapEnv_NilInterfaceValue(t *testing.T) {
	env := map[string]any{"a": nil}
	c := &nature.Cache{}

	got := conf.EnvWithCache(c, env)

	field, ok := got.Fields["a"]
	if !ok {
		t.Fatalf("Fields[%q] missing", "a")
	}
	if !field.Nil {
		t.Errorf("Fields[%q].Nil = %v, want true", "a", field.Nil)
	}
}

// TestEnvWithCache_TypesMapEnv_Delegates verifies that a types.Map env is
// delegated to its own Nature() method, i.e. EnvWithCache returns exactly
// what env.Nature() would produce.
//
//loop:behavior envwithcache-envwithcache-delegates-types-map-env-to-its-own
func TestEnvWithCache_TypesMapEnv_Delegates(t *testing.T) {
	env := types.Map{"x": types.Int}
	c := &nature.Cache{}

	got := conf.EnvWithCache(c, env)
	want := env.Nature()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("EnvWithCache(c, env) = %#v, want env.Nature() = %#v", got, want)
	}
}

// TestEnvWithCache_UnsupportedKind_Panics verifies that env values which are
// neither nil, a types.Map, a struct, nor a map cause EnvWithCache to panic
// with a "unknown type %T" message.
//
//loop:behavior envwithcache-envwithcache-panics-for-env-kinds-other-than-str
func TestEnvWithCache_UnsupportedKind_Panics(t *testing.T) {
	tests := []struct {
		name string
		env  any
	}{
		{name: "int", env: 42},
		{name: "string", env: "s"},
		{name: "slice", env: []int{1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &nature.Cache{}
			want := fmt.Sprintf("unknown type %T", tt.env)

			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("EnvWithCache(c, %#v) did not panic, want panic %q", tt.env, want)
				}
				got, ok := r.(string)
				if !ok {
					t.Fatalf("panic value = %#v (%T), want string %q", r, r, want)
				}
				if got != want {
					t.Errorf("panic value = %q, want %q", got, want)
				}
			}()

			conf.EnvWithCache(c, tt.env)
		})
	}
}
