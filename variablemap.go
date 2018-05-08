package main

import (
	"reflect"
)

type VariableMap map[string]interface{}

// this function assumes to and from are both maps
// checks to ensure that key/values from from can fit into to
func mapsCompatible(to, from reflect.Value) bool {
	return from.Type().Key().AssignableTo(to.Type().Key())
}

const mergeBadKeysPanic = "cannot merge conflicting map key types"
const mergeBadValsPanic = "cannot merge conflicting map value types"

func mergeReflectingMaps(to, from reflect.Value) {
	if !mapsCompatible(to, from) {
		panic(mergeBadKeysPanic)
	}

	for _, k := range from.MapKeys() {
		toVal := to.MapIndex(k)
		if !toVal.IsValid() {
			fromVal := from.MapIndex(k)
			if !fromVal.Type().AssignableTo(to.Type().Elem()) {
				panic(mergeBadValsPanic)
			}

			to.SetMapIndex(k, fromVal)
			continue
		}

		toValReeval := reflect.ValueOf(toVal.Interface())
		if toValReeval.Type().Kind() == reflect.Map {
			fromVal := from.MapIndex(k)
			fromValReeval := reflect.ValueOf(fromVal.Interface())
			if fromValReeval.Type().Kind() == reflect.Map {
				// we must go deeper!
				mergeReflectingMaps(toValReeval, fromValReeval)
			}
		}
	}
}

func (v VariableMap) MergeFrom(v2 VariableMap) VariableMap {
	for k, newVal := range v2 {
		existingVal, h := v[k]
		if !h {
			v[k] = newVal
			// quick exit before we try to do reflection
			continue
		}

		// special case where both are maps - we will try to merge lower!
		reflNewVal := reflect.ValueOf(newVal)
		if reflNewVal.Type().Kind() == reflect.Map {
			// this is a map
			reflExistingVal := reflect.ValueOf(existingVal)
			if reflExistingVal.Type().Kind() == reflect.Map {
				mergeReflectingMaps(reflExistingVal, reflNewVal)
			}
		}
	}
	return v
}
