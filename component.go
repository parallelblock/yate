package main

import (
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
)

type RenderContext interface {
	Resolve(base, relative string) (string, error)
	Render(componentName string, args ...interface{}) (err error)
	Writer() io.Writer
	Vars() interface{}
}

type FilesystemLoader interface {
	Load(path string) (string, error)
}

type Component struct {
	*template.Template
	Filepath string

	ctxMx sync.Locker
	ctx   RenderContext
}

func NewComponent(filepath string) *Component {
	c := new(Component)
	f := template.FuncMap{
		"include": func(component string, fargs ...interface{}) (interface{}, error) {
			componentPath, err := c.ctx.Resolve(c.Filepath, component)
			if err != nil {
				return "", err
			}

			err = c.ctx.Render(componentPath, fargs...)
			return "", err
		},
	}

	c.Template = template.New("").Funcs(f)
	c.Filepath = filepath
	c.ctxMx = new(sync.Mutex)
	return c
}

func (c *Component) Render(ctx RenderContext, args ...interface{}) (err error) {
	v := make(map[string]interface{})
	v["Vars"] = ctx.Vars()
	for i, arg := range args {
		v["Arg"+strconv.Itoa(i)] = arg
	}

	c.ctxMx.Lock()
	defer c.ctxMx.Unlock()
	c.ctx = ctx
	err = c.Execute(ctx.Writer(), v)
	return
}

type CyclicalRenderDependenciesError struct {
	Stack []string
}

func (c *CyclicalRenderDependenciesError) Error() string {
	return "Cycle detected in include path: " + strings.Join(c.Stack, " -> ")
}

type ComponentResolver interface {
	Resolve(path string) (*Component, error)
}

type RenderScope struct {
	W           io.Writer
	Resolver    ComponentResolver
	BasePath    string
	Variables   interface{}
	RenderStack []*Component
}

func NewRenderScope(w io.Writer, r ComponentResolver, basePath string, vars interface{}) *RenderScope {
	return &RenderScope{
		W:           w,
		Resolver:    r,
		BasePath:    basePath,
		Variables:   vars,
		RenderStack: make([]*Component, 0),
	}
}

func (c *RenderScope) Resolve(base, relative string) (string, error) {
	dir, _ := filepath.Split(base)

	return filepath.Rel(c.BasePath, filepath.Join(dir, relative))
}

func (c *RenderScope) Render(componentName string, args ...interface{}) (err error) {
	comp, err := c.Resolver.Resolve(componentName)
	if err != nil {
		return err
	}

	for _, v := range c.RenderStack {
		if v == comp {
			slen := len(c.RenderStack) + 1
			stack := make([]string, slen)
			for i, v := range c.RenderStack {
				stack[i] = v.Filepath
			}
			stack[slen-1] = comp.Filepath
			return &CyclicalRenderDependenciesError{
				Stack: stack,
			}
		}
	}

	c.RenderStack = append(c.RenderStack, comp)

	return comp.Render(c, args...)
}

func (c *RenderScope) Vars() interface{} {
	return c.Variables
}

func (c *RenderScope) Writer() io.Writer {
	return c.W
}

type FileReader func(filename string) ([]byte, error)

type CacheComponentResolver struct {
	sync.Mutex
	Downstream FileReader
	cache      map[string]*Component
}

func NewCacheComponentResolver(downstream FileReader) *CacheComponentResolver {
	return &CacheComponentResolver{
		Downstream: downstream,
		cache:      make(map[string]*Component),
	}
}

func (r *CacheComponentResolver) Resolve(path string) (*Component, error) {
	r.Lock()
	defer r.Unlock()

	c, h := r.cache[path]
	if !h {
		b, e := r.Downstream(path)
		if e != nil {
			return nil, e
		}
		c = NewComponent(path)
		c.Parse(string(b))
		r.cache[path] = c
	}
	return c, nil
}

type TrackingComponentResolver struct {
	Downstream ComponentResolver
	hits       map[string]struct{}
}

func NewTrackingComponentResolver(downstream ComponentResolver) *TrackingComponentResolver {
	return &TrackingComponentResolver{
		Downstream: downstream,
		hits:       make(map[string]struct{}),
	}
}

func (r *TrackingComponentResolver) Resolve(path string) (c *Component, e error) {
	c, e = r.Downstream.Resolve(path)
	if e != nil {
		return
	}
	r.hits[path] = struct{}{}
	return
}

func (r *TrackingComponentResolver) Hits() map[string]struct{} {
	return r.hits
}
