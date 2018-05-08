package main

type ResourceConfig struct {
	Template  string
	Output    string
	Variables VariableMap
}

type ResourceConfigSource interface {
	GlobalVariables() VariableMap
	GetConfig(resource string) ResourceConfig
}

type Resource struct {
	Name string
}
