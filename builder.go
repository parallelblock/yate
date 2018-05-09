package main

type ResourceConfig struct {
	Template  string
	Output    string
	Inherits  []string
	Variables VariableMap
}

func stringArrayIs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func (r *ResourceConfig) Is(r2 *ResourceConfig) bool {
	if r2 == nil {
		return false
	} else if r == r2 {
		return true
	} else if r.Template != r2.Template {
		return false
	} else if r.Output != r2.Output {
		return false
	} else if !stringArrayIs(r, r2) {
		return false
	}
}

type ResourceConfigSource interface {
	GlobalVariables() VariableMap
	GetConfig(resource string) ResourceConfig
}

type Resource struct {
	Name string
}
