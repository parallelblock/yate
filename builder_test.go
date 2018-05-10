package main

import (
	"strconv"
	"testing"
)

var resourceConfigEqualTests = []struct {
	a, b     ResourceConfig
	response bool
}{
	{
		ResourceConfig{
			"a",
			"b",
			[]string{},
			VariableMap{},
		},
		ResourceConfig{
			"a",
			"b",
			[]string{},
			VariableMap{},
		},
		true,
	},
	{
		ResourceConfig{
			"a",
			"b",
			[]string{"a"},
			VariableMap{},
		},
		ResourceConfig{
			"a",
			"b",
			[]string{"a"},
			VariableMap{},
		},
		true,
	},
	{
		ResourceConfig{
			"a",
			"b",
			[]string{"a"},
			VariableMap{"a": "a"},
		},
		ResourceConfig{
			"a",
			"b",
			[]string{"a"},
			VariableMap{"a": "a"},
		},
		true,
	},
	{
		ResourceConfig{
			"a",
			"b",
			[]string{"a"},
			VariableMap{"a": map[string]string{"a": "a"}},
		},
		ResourceConfig{
			"a",
			"b",
			[]string{"a"},
			VariableMap{"a": map[string]string{"a": "a"}},
		},
		true,
	},
	{
		ResourceConfig{
			"a",
			"b",
			[]string{"a"},
			VariableMap{"a": map[string]string{"a": "a"}},
		},
		ResourceConfig{
			"b",
			"b",
			[]string{"a"},
			VariableMap{"a": map[string]string{"a": "a"}},
		},
		false,
	},
	{
		ResourceConfig{
			"a",
			"b",
			[]string{"a"},
			VariableMap{"a": map[string]string{"a": "a"}},
		},
		ResourceConfig{
			"a",
			"c",
			[]string{"a"},
			VariableMap{"a": map[string]string{"a": "a"}},
		},
		false,
	},
	{
		ResourceConfig{
			"a",
			"b",
			[]string{},
			VariableMap{"a": map[string]string{"a": "a"}},
		},
		ResourceConfig{
			"a",
			"b",
			[]string{"a"},
			VariableMap{"a": map[string]string{"a": "a"}},
		},
		false,
	},
	{
		ResourceConfig{
			"a",
			"b",
			[]string{"a"},
			VariableMap{"a": map[string]string{"a": "a"}},
		},
		ResourceConfig{
			"a",
			"b",
			[]string{},
			VariableMap{"a": map[string]string{"a": "a"}},
		},
		false,
	},
	{
		ResourceConfig{
			"a",
			"b",
			[]string{"a"},
			VariableMap{"a": map[string]string{"a": "a"}},
		},
		ResourceConfig{
			"a",
			"b",
			[]string{"b"},
			VariableMap{"a": map[string]string{"a": "a"}},
		},
		false,
	},
	{
		ResourceConfig{
			"a",
			"b",
			[]string{"a", "b"},
			VariableMap{"a": map[string]string{"a": "a"}},
		},
		ResourceConfig{
			"a",
			"b",
			[]string{"b", "a"},
			VariableMap{"a": map[string]string{"a": "a"}},
		},
		false,
	},
	{
		ResourceConfig{
			"a",
			"b",
			[]string{},
			VariableMap{"a": map[string]string{"a": "a"}},
		},
		ResourceConfig{
			"a",
			"b",
			[]string{},
			VariableMap{"a": map[string]string{"a": "b"}},
		},
		false,
	},
	{
		ResourceConfig{
			"a",
			"b",
			[]string{},
			VariableMap{"a": "a"},
		},
		ResourceConfig{
			"a",
			"b",
			[]string{},
			VariableMap{"a": map[string]string{"a": "b"}},
		},
		false,
	},
	{
		ResourceConfig{
			"a",
			"b",
			[]string{},
			VariableMap{"a": map[string]string{"a": "b"}},
		},
		ResourceConfig{
			"a",
			"b",
			[]string{},
			VariableMap{"a": "a"},
		},
		false,
	},
}

func TestResourceConfigIs(t *testing.T) {
	for i, tt := range resourceConfigEqualTests {
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			is := (&tt.a).Is(&tt.b)
			if is != tt.response {
				t.Errorf("expected %v, got %v", tt.response, is)
			}
		})
	}
}

func TestResourceConfigIsNilFails(t *testing.T) {
	a := &ResourceConfig{}
	is := a.Is(nil)
	if is {
		t.Errorf("nil is response to non-nil value is true")
	}
}

func TestResourceConfigIsSamePasses(t *testing.T) {
	a := &ResourceConfig{}
	is := a.Is(a)
	if !is {
		t.Errorf("equality against self failed")
	}
}
