// Package templater is imported form https://github.com/go-task/task/tree/master/internal/templater
package templater

import (
	"bytes"
	"text/template"
)

type (
	// A Templater is the interface that wraps basic templating methods that can be called several
	// times, without having to check for error each time. The first error that
	// happen will be assigned to the rhe inner struct, and consecutive calls to funcs will just
	// return the zero value.
	Templater interface {
		Replace(string) string
		ReplaceSlice([]string) []string
		ReplaceMap(map[string]string) map[string]string
		ReplaceMapI(map[string]any) map[string]any
		Err() error
	}

	// A Context carries the context of a Templater.
	Context interface {
		Variables() map[string]string
	}

	templater struct {
		ctx Context
		err error
	}
)

// New returns a new Templater.
func New(ctx Context) Templater {
	return &templater{
		ctx: ctx,
	}
}

func (r *templater) Replace(str string) string {
	if r.err != nil || str == "" {
		return ""
	}

	templ, err := template.New("").Funcs(templateFuncs).Parse(str)
	if err != nil {
		r.err = err
		return ""
	}

	var b bytes.Buffer
	if err = templ.Execute(&b, r.ctx.Variables()); err != nil {
		r.err = err
		return ""
	}
	return b.String()
}

func (r *templater) ReplaceSlice(strs []string) []string {
	if r.err != nil || len(strs) == 0 {
		return nil
	}

	new := make([]string, len(strs))
	for i, str := range strs {
		new[i] = r.Replace(str)
	}
	return new
}

func (r *templater) ReplaceMap(m map[string]string) map[string]string {
	if r.err != nil || len(m) == 0 {
		return nil
	}

	new := make(map[string]string, len(m))
	for k, v := range m {
		new[k] = r.Replace(v)
	}
	return new
}

func (r *templater) ReplaceMapI(m map[string]any) map[string]any {
	if r.err != nil || len(m) == 0 {
		return nil
	}

	new := make(map[string]any, len(m))
	for k, v := range m {
		if s, ok := v.(string); ok {
			new[k] = r.Replace(s)
			continue
		}
		new[k] = v
	}
	return new
}

func (r *templater) Err() error {
	return r.err
}
