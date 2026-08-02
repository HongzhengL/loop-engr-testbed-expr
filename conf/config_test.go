package conf_test

import (
	"reflect"
	"testing"

	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/builtin"
	"github.com/expr-lang/expr/conf"
)

// TestCreateNew_OptimizeAndShortCircuitDefaults verifies that CreateNew
// enables Optimize and ShortCircuit by default.
//
//loop:behavior createnew-createnew-enables-optimize-and-shortcircuit-by-d
func TestCreateNew_OptimizeAndShortCircuitDefaults(t *testing.T) {
	c := conf.CreateNew()

	if !c.Optimize {
		t.Errorf("Optimize = %v, want true", c.Optimize)
	}
	if !c.ShortCircuit {
		t.Errorf("ShortCircuit = %v, want true", c.ShortCircuit)
	}
}

// TestCreateNew_MaxNodesDefault verifies that CreateNew sets MaxNodes to the
// current value of the package var DefaultMaxNodes.
//
//loop:behavior createnew-createnew-sets-maxnodes-to-defaultmaxnodes
func TestCreateNew_MaxNodesDefault(t *testing.T) {
	want := conf.DefaultMaxNodes

	c := conf.CreateNew()

	if c.MaxNodes != want {
		t.Errorf("MaxNodes = %v, want %v (conf.DefaultMaxNodes)", c.MaxNodes, want)
	}
}

// TestCreateNew_InitializesEmptyMaps verifies that CreateNew initializes
// ConstFns, Functions and Disabled as non-nil, empty maps.
//
//loop:behavior createnew-createnew-initializes-constfns-functions-disable
func TestCreateNew_InitializesEmptyMaps(t *testing.T) {
	c := conf.CreateNew()

	if c.ConstFns == nil {
		t.Error("ConstFns is nil, want non-nil")
	} else if len(c.ConstFns) != 0 {
		t.Errorf("len(ConstFns) = %d, want 0", len(c.ConstFns))
	}

	if c.Functions == nil {
		t.Error("Functions is nil, want non-nil")
	} else if len(c.Functions) != 0 {
		t.Errorf("len(Functions) = %d, want 0", len(c.Functions))
	}

	if c.Disabled == nil {
		t.Error("Disabled is nil, want non-nil")
	} else if len(c.Disabled) != 0 {
		t.Errorf("len(Disabled) = %d, want 0", len(c.Disabled))
	}
}

// TestCreateNew_PopulatesBuiltins verifies that CreateNew populates Builtins
// from builtin.Builtins, keyed by function name.
//
//loop:behavior createnew-createnew-populates-builtins-from-builtin-builti
func TestCreateNew_PopulatesBuiltins(t *testing.T) {
	c := conf.CreateNew()

	if c.Builtins == nil {
		t.Fatal("Builtins is nil, want non-nil")
	}
	if len(c.Builtins) != len(builtin.Builtins) {
		t.Fatalf("len(Builtins) = %d, want %d (len(builtin.Builtins))", len(c.Builtins), len(builtin.Builtins))
	}
	for _, f := range builtin.Builtins {
		got, ok := c.Builtins[f.Name]
		if !ok {
			t.Errorf("Builtins[%q] missing", f.Name)
			continue
		}
		if got != f {
			t.Errorf("Builtins[%q] = %p, want %p (same *builtin.Function)", f.Name, got, f)
		}
	}
}

// TestNew_AppliesEnvOnTopOfDefaults verifies that New applies the given env
// while keeping the same defaults CreateNew would set.
//
//loop:behavior new-new-applies-given-env-on-top-of-createnew-defaul
func TestNew_AppliesEnvOnTopOfDefaults(t *testing.T) {
	env := struct{ X int }{}

	c := conf.New(env)

	if !c.Optimize {
		t.Errorf("Optimize = %v, want true", c.Optimize)
	}
	if !c.ShortCircuit {
		t.Errorf("ShortCircuit = %v, want true", c.ShortCircuit)
	}
	if c.MaxNodes != conf.DefaultMaxNodes {
		t.Errorf("MaxNodes = %v, want %v (conf.DefaultMaxNodes)", c.MaxNodes, conf.DefaultMaxNodes)
	}
	if c.EnvObject != any(env) {
		t.Errorf("EnvObject = %#v, want %#v", c.EnvObject, env)
	}
}

// TestConfig_WithEnv_SetsEnvObject verifies that WithEnv sets EnvObject to
// the given env.
//
//loop:behavior config-withenv-withenv-sets-envobject-to-given-env
func TestConfig_WithEnv_SetsEnvObject(t *testing.T) {
	c := conf.CreateNew()
	env := struct{ X int }{X: 1}

	c.WithEnv(env)

	if c.EnvObject != any(env) {
		t.Errorf("EnvObject = %#v, want %#v", c.EnvObject, env)
	}
}

// TestConfig_WithEnv_PropagatesStrict verifies that WithEnv propagates the
// Strict flag from the computed Env into Config.Strict.
//
//loop:behavior config-withenv-withenv-propagates-env-strict-into-config-strict
func TestConfig_WithEnv_PropagatesStrict(t *testing.T) {
	c := conf.CreateNew()
	env := struct{ X int }{}

	c.WithEnv(env)

	if !c.Env.Strict {
		t.Fatalf("Env.Strict = %v, want true (struct env is marked strict)", c.Env.Strict)
	}
	if c.Strict != c.Env.Strict {
		t.Errorf("Strict = %v, want %v (equal to Env.Strict)", c.Strict, c.Env.Strict)
	}
}

// TestConfig_ConstExpr_PanicsWithoutEnv verifies that ConstExpr panics when
// no environment is set on the Config.
//
//loop:behavior config-constexpr-constexpr-panics-when-no-environment-is-set
func TestConfig_ConstExpr_PanicsWithoutEnv(t *testing.T) {
	c := conf.CreateNew()
	want := "no environment is specified for ConstExpr()"

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("ConstExpr did not panic, want panic %q", want)
		}
		got, ok := r.(string)
		if !ok {
			t.Fatalf("panic value = %#v (%T), want string %q", r, r, want)
		}
		if got != want {
			t.Errorf("panic value = %q, want %q", got, want)
		}
	}()

	c.ConstExpr("anything")
}

// TestConfig_ConstExpr_PanicsWhenMemberNotFunction verifies that ConstExpr
// panics with a descriptive error when the named env member is not a
// function.
//
//loop:behavior config-constexpr-constexpr-panics-when-named-member-is-not-a-func
func TestConfig_ConstExpr_PanicsWhenMemberNotFunction(t *testing.T) {
	c := conf.CreateNew()
	env := map[string]any{"n": 42}
	c.WithEnv(env)

	want := `const expression "n" must be a function`

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("ConstExpr did not panic, want panic %q", want)
		}
		err, ok := r.(error)
		if !ok {
			t.Fatalf("panic value = %#v (%T), want an error with message %q", r, r, want)
		}
		if err.Error() != want {
			t.Errorf("panic error message = %q, want %q", err.Error(), want)
		}
	}()

	c.ConstExpr("n")
}

// TestConfig_ConstExpr_RegistersFunctionMember verifies that ConstExpr
// registers a func-typed env member into ConstFns without panicking.
//
//loop:behavior config-constexpr-constexpr-registers-function-member-into-constfn
func TestConfig_ConstExpr_RegistersFunctionMember(t *testing.T) {
	c := conf.CreateNew()
	env := map[string]any{"f": func() int { return 1 }}
	c.WithEnv(env)

	c.ConstExpr("f")

	fn, ok := c.ConstFns["f"]
	if !ok {
		t.Fatalf("ConstFns[%q] missing after ConstExpr", "f")
	}
	if fn.Kind() != reflect.Func {
		t.Errorf("ConstFns[%q].Kind() = %v, want %v", "f", fn.Kind(), reflect.Func)
	}
}

// TestConfig_IsOverridden_TrueInFunctions verifies that IsOverridden
// returns true when the name is registered in Functions.
//
//loop:behavior config-isoverridden-isoverridden-true-when-name-is-in-functions
func TestConfig_IsOverridden_TrueInFunctions(t *testing.T) {
	c := conf.CreateNew()
	c.Functions["myFunc"] = &builtin.Function{Name: "myFunc"}

	if !c.IsOverridden("myFunc") {
		t.Errorf("IsOverridden(%q) = false, want true", "myFunc")
	}
}

// TestConfig_IsOverridden_TrueInEnv verifies that IsOverridden returns true
// when the name resolves through Env, even without a Functions entry.
//
//loop:behavior config-isoverridden-isoverridden-true-when-name-resolves-in-env
func TestConfig_IsOverridden_TrueInEnv(t *testing.T) {
	c := conf.CreateNew()
	env := map[string]any{"bar": 1}
	c.WithEnv(env)

	if len(c.Functions) != 0 {
		t.Fatalf("precondition failed: Functions is not empty: %v", c.Functions)
	}
	if !c.IsOverridden("bar") {
		t.Errorf("IsOverridden(%q) = false, want true", "bar")
	}
}

// TestConfig_IsOverridden_FalseForUnknownName verifies that IsOverridden
// returns false when the name is neither in Functions nor resolvable
// through Env.
//
//loop:behavior config-isoverridden-isoverridden-false-for-unknown-name
func TestConfig_IsOverridden_FalseForUnknownName(t *testing.T) {
	c := conf.CreateNew()

	if c.IsOverridden("doesNotExist") {
		t.Errorf("IsOverridden(%q) = true, want false", "doesNotExist")
	}
}

// checkingVisitor is a minimal ast.Visitor that also implements
// conf.Checker, recording whether Check was invoked.
type checkingVisitor struct {
	checked bool
}

func (v *checkingVisitor) Visit(node *ast.Node) {}

func (v *checkingVisitor) Check() {
	v.checked = true
}

// TestConfig_Check_InvokesCheckOnVisitors verifies that Config.Check invokes
// Check() on every visitor that implements the conf.Checker interface.
//
//loop:behavior config-check-check-invokes-check-on-visitors-implementing-che
func TestConfig_Check_InvokesCheckOnVisitors(t *testing.T) {
	c := conf.CreateNew()
	v := &checkingVisitor{}
	c.Visitors = append(c.Visitors, v)

	c.Check()

	if !v.checked {
		t.Errorf("checkingVisitor.Check() was not invoked, want it to be invoked at least once")
	}
}
