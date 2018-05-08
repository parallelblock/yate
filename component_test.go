package main

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
	"testing"
)

type rfunc func(args []interface{}) error

type testRCtx struct {
	b            *bytes.Buffer
	subrender    map[string]rfunc
	vars         interface{}
	resolveFails bool
}

var notExist = errors.New("does not exist")
var resolveFail = errors.New("resolve failure")

func (c *testRCtx) Resolve(base, relative string) (string, error) {
	if c.resolveFails {
		return "", resolveFail
	}
	return relative, nil
}

func (c *testRCtx) Render(componentName string, args ...interface{}) error {
	f, h := c.subrender[componentName]
	if !h {
		return notExist
	}
	return f(args)
}

func (c *testRCtx) Writer() io.Writer {
	return c.b
}

func (c *testRCtx) Vars() interface{} {
	return c.vars
}

var componentTests = []struct {
	template     string
	resolveFails bool
	result       string
	resultError  error
}{
	{"hello world", false, "hello world", nil},                           // simple
	{"{{ .Vars.test }}", false, "hello!", nil},                           // var replacement
	{"{{ .Arg0 }}, {{ .Arg1 }}", false, "the 0th arg, the 1st arg", nil}, // arg replacement
	{"{{ include \"something\" }}", false, "something?", nil},            // inclusion of render
	{"{{ include \"subargs\" 123 }}", false, "subargs test", nil},        // inclusion with specific arguments
	{"{{ include \"nothing\" }}", true, "", resolveFail},                 // resolution failure
	{"{{ include \"nonexistant\" }}", false, "", notExist},               // rendering failure
}

func TestComponent(t *testing.T) {
	for _, tt := range componentTests {
		t.Run(tt.template, func(t *testing.T) {
			var rctx *testRCtx
			rctx = &testRCtx{
				b: new(bytes.Buffer),
				subrender: map[string]rfunc{
					"something": func(_ []interface{}) error {
						_, e := rctx.b.WriteString("something?")
						return e
					},
					"subargs": func(args []interface{}) error {
						if len(args) != 1 {
							t.Fatalf("incorrect arg length - expected: %d, got: %d", 1, len(args))
						}
						if args[0] != 123 {
							t.Fatalf("incorrect arg contents - expected: %v, got %v", 123, args[0])
						}
						_, e := rctx.b.WriteString("subargs test")
						return e
					},
				},
				vars: map[string]string{
					"test": "hello!",
				},
				resolveFails: tt.resolveFails,
			}
			c := NewComponent("base")

			_, err := c.Parse(tt.template)
			if err != nil {
				t.Fatal("template compile failed", err)
			}

			err = c.Render(rctx, "the 0th arg", "the 1st arg")
			if err == nil && tt.resultError != nil {
				t.Fatalf("template returned unexpected error - expected: %v, got: %v", tt.resultError, err)
			} else if err != nil && !strings.HasSuffix(err.Error(), tt.resultError.Error()) {
				t.Fatalf("template returned unexpected error - expected: %v, got: %v", tt.resultError, err)
			}

			if err == nil && tt.result != rctx.b.String() {
				t.Fatalf("template generated incorrect result - expected: %s, got: %s", tt.result, rctx.b.String())
			}
		})
	}
}

var renderScopeResolveTests = []struct {
	callBase, renderBase, resolve, result string
}{
	{"src/test.file", "./", "something.file", "src/something.file"},
	{"src/test.file", "./", "../something.file", "something.file"},
	{"src/test.file", "./", "something/something.file", "src/something/something.file"},
	{"src/test.file", "src/", "something.file", "something.file"},
}

func TestRenderScopeResolve(t *testing.T) {
	for _, tt := range renderScopeResolveTests {
		t.Run(tt.renderBase+"$"+tt.callBase, func(t *testing.T) {
			r := RenderScope{
				BasePath: tt.renderBase,
			}
			s, e := r.Resolve(tt.callBase, tt.resolve)
			if e != nil {
				t.Fatalf("unexpected error: %v", e)
			}

			if s != tt.result {
				t.Fatalf("incorrect result - expected: %s, got %s", s, tt.result)
			}
		})
	}
}

type staticResolver map[string]string

func (s staticResolver) Resolve(path string) (*Component, error) {
	v, h := s[path]
	if !h {
		return nil, notExist
	}
	c := NewComponent(path)
	_, err := c.Option("missingkey=error").Parse(v)
	if err != nil {
		panic(err)
	}
	return c, nil
}

var renderScopeRenderTests = []struct {
	resolver      staticResolver
	componentName string
	args          []interface{}
	expected      string
}{
	{staticResolver{"a": "test"}, "a", []interface{}{}, "test"},                                                     // basic functionality
	{staticResolver{"a": "{{ .Arg0 }}, {{ .Arg1 }}"}, "a", []interface{}{"1st arg", "2nd arg"}, "1st arg, 2nd arg"}, // args passthrough
	{staticResolver{"a": "{{ include \"b\" }}", "b": "test"}, "a", []interface{}{}, "test"},                         // include stack
	{staticResolver{"a": "{{ .Vars.v }}"}, "a", []interface{}{}, "result"},                                          // vars passthrough
}

func TestRenderScopeRender(t *testing.T) {
	for i, tt := range renderScopeRenderTests {
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			b := new(bytes.Buffer)
			scope := NewRenderScope(b, tt.resolver, "", map[string]string{"v": "result"})

			t.Log(tt.args)

			e := scope.Render(tt.componentName, tt.args...)
			if e != nil {
				t.Fatalf("unexpected render error: %s", e)
			}

			if b.String() != tt.expected {
				t.Fatalf("unexpected result - expected: %s, got: %s", tt.expected, b.String())
			}
		})
	}
}

type absoluteResolver map[string]*Component

func (s absoluteResolver) Resolve(path string) (*Component, error) {
	return s[path], nil
}

func TestRenderScopeRenderRecursiveErrors(t *testing.T) {
	r := make(absoluteResolver)
	ac := NewComponent("a")
	ac.Parse("{{ include \"b\" }}")
	bc := NewComponent("b")
	bc.Parse("{{ include \"a\" }}")

	r["a"] = ac
	r["b"] = bc

	b := new(bytes.Buffer)
	scope := NewRenderScope(b, r, "", struct{}{})
	e := scope.Render("a")
	if e == nil {
		t.Fatalf("expected an error, got nil")
	}

	err := "template: :1:3: executing \"\" at <include \"b\">: error calling include: template: :1:3: executing \"\" at <include \"a\">: error calling include: Cycle detected in include path: a -> b -> a"

	if e.Error() != err {
		t.Fatalf("unexpected error result: expected %s, got %s", err, e.Error())
	}
}

type faultyResolver struct{}

func (s faultyResolver) Resolve(path string) (*Component, error) {
	return nil, notExist
}

func TestRenderScopeRenderFaultyResolverErrors(t *testing.T) {
	r := faultyResolver{}
	b := new(bytes.Buffer)
	scope := NewRenderScope(b, r, "", struct{}{})
	e := scope.Render("a")
	if e != notExist {
		t.Fatalf("error was not correct - expected %v, got %v", notExist, e)
	}
}

func TestCacheComponentResolver(t *testing.T) {
	shouldResolve := true
	ccr := NewCacheComponentResolver(func(filename string) ([]byte, error) {
		if !shouldResolve {
			return []byte{}, notExist
		}

		switch filename {
		case "a":
			return []byte{0x48}, nil
		case "b":
			return []byte{0x49}, nil
		}
		return []byte{}, notExist
	})

	errChk := func(i *Component, e error) {
		if i == nil {
			t.Fatalf("unexpected nil value")
		}
		if e != nil {
			t.Fatalf("unexpected error: %s", e)
		}
	}

	a, err := ccr.Resolve("a")
	errChk(a, err)

	if a.Filepath != "a" {
		t.Errorf("incorrect filepath for a - expected %s, got %s", "a", a.Filepath)
	}

	_, err = ccr.Resolve("c")
	if err != notExist {
		t.Errorf("incorrect error - expected: %s, got %s", notExist, err)
	}

	b, err := ccr.Resolve("b")
	errChk(b, err)

	if b.Filepath != "b" {
		t.Errorf("incorrect filepath for b - expected %s, got %s", "b", b.Filepath)
	}

	shouldResolve = false

	a2, err := ccr.Resolve("a")
	errChk(a, err)

	if a != a2 {
		t.Errorf("was not given back exact duplicate for dual resolution of a")
	}
}

var trackingComponentTests = []struct {
	query []string
	hits  []string
}{
	{[]string{"a"}, []string{"a"}},
	{[]string{"a", "b"}, []string{"a", "b"}},
	{[]string{"a", "e", "a"}, []string{"a"}},
	{[]string{"e"}, []string{}},
}

func in(test string, array []string) bool {
	for _, v := range array {
		if test == v {
			return true
		}
	}
	return false
}

func TestTrackingComponentResolver(t *testing.T) {
	d := staticResolver{
		"a": "a",
		"b": "b",
		"c": "c",
	}

	for _, tt := range trackingComponentTests {
		t.Run(strings.Join(tt.query, "-"), func(t *testing.T) {
			r := NewTrackingComponentResolver(d)
			for _, q := range tt.query {
				_, e := r.Resolve(q)
				if in(q, tt.hits) {
					if e != nil {
						t.Errorf("query with hit returned error: %s", e)
					}
				} else {
					if e != notExist {
						t.Errorf("unexpected error returned - expected %s, got %s", notExist, e)
					}
				}
			}

			for _, h := range tt.hits {
				_, x := r.Hits()[h]
				if !x {
					t.Errorf("expected hit %s, but was not present.", h)
				}
			}

			for h, _ := range r.Hits() {
				if !in(h, tt.hits) {
					t.Errorf("found hit %s, but was not present in hitlist.", h)
				}
			}
		})
	}
}
