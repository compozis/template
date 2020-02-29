package template

import (
	"html/template"
	"sync"
)

// Cache defines interface for prepared template caching for Engine.
//
// All Cache implementations must be thread-safe.
type Cache interface {
	get(name string) *template.Template
	put(name string, tmpl *template.Template)
}

// NoCache implements Cache without any caching mechanism.
type NoCache struct{}

func (n NoCache) get(name string) *template.Template {
	return nil
}

func (n NoCache) put(name string, tmpl *template.Template) {
}

// PermanentCache implements Cache to store prepared templates permanently.
type PermanentCache struct {
	templates map[string]*template.Template

	mux sync.RWMutex
}

func NewPermanentCache() *PermanentCache {
	return &PermanentCache{
		templates: map[string]*template.Template{},
	}
}

func (p *PermanentCache) get(name string) *template.Template {
	p.mux.RLock()
	tmpl, found := p.templates[name]
	p.mux.RUnlock()

	if !found {
		tmpl = nil
	}
	return tmpl
}

func (p *PermanentCache) put(name string, tmpl *template.Template) {
	p.mux.Lock()
	p.templates[name] = tmpl
	p.mux.Unlock()
}
