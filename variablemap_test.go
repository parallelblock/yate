package main

import (
	"reflect"
	"strconv"
	"testing"
)

var variableMapMergeTests = []struct {
	object, argument, result VariableMap
}{
	{VariableMap{"a": "value1"}, VariableMap{"b": "value2"}, VariableMap{"a": "value1", "b": "value2"}},
	{VariableMap{"a": "value1"}, VariableMap{"a": "value2"}, VariableMap{"a": "value1"}},
	{VariableMap{"a": map[string]string{"a": "value1"}}, VariableMap{"a": map[string]string{"b": "value2"}}, VariableMap{"a": map[string]string{"a": "value1", "b": "value2"}}},
	{VariableMap{"a": "value1"}, VariableMap{"a": map[string]string{"b": "value2"}}, VariableMap{"a": "value1"}},
	{VariableMap{"a": map[string]string{"a": "value1"}}, VariableMap{"a": map[string]string{"a": "value2"}}, VariableMap{"a": map[string]string{"a": "value1"}}},
	{VariableMap{"a": map[string]interface{}{"a": map[string]string{"a": "value1"}}}, VariableMap{"a": map[string]string{"a": "value2"}}, VariableMap{"a": map[string]interface{}{"a": map[string]string{"a": "value1"}}}},
	{VariableMap{"a": map[string]interface{}{"a": map[string]string{"a": "value1"}}}, VariableMap{"a": map[string]interface{}{"a": map[string]string{"b": "value2"}}}, VariableMap{"a": map[string]interface{}{"a": map[string]string{"a": "value1", "b": "value2"}}}},
}

func TestVariableMapMerge(t *testing.T) {
	for i, tt := range variableMapMergeTests {
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			result := tt.object.MergeFrom(tt.argument)
			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("unequal results - expected %v, got %v", tt.result, result)
			}
		})
	}
}

func TestOddballMapMergeKeys(t *testing.T) {
	defer func() {
		r := recover()
		if r != mergeBadKeysPanic {
			t.Errorf("failed with incorrect panic - expected %s got %s", mergeBadKeysPanic, r)
		}
	}()

	a := VariableMap{"a": map[string]string{"a": "a"}}
	b := VariableMap{"a": map[int]string{3: "maybe"}}
	a.MergeFrom(b)

	t.Errorf("did not panic")
}

func TestOddballMapMergeValues(t *testing.T) {
	defer func() {
		r := recover()
		if r != mergeBadValsPanic {
			t.Errorf("failed with incorrect panic - expected %s got %s", mergeBadValsPanic, r)
		}
	}()

	a := VariableMap{"a": map[string]string{"a": "a"}}
	b := VariableMap{"a": map[string]int{"b": 3}}
	a.MergeFrom(b)

	t.Errorf("did not panic")
}
