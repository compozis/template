package template

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"regexp"
	"sync"
)

type FuncMap template.FuncMap

type Engine struct {
	fs FileSystem

	modifyMux     sync.Mutex
	extendsRegexp *regexp.Regexp
	funcMap       FuncMap
	cache         Cache
}

func NewEngine(fs FileSystem) *Engine {
	extendsRegex := buildExtendsRegexp()

	return &Engine{
		fs: fs,

		extendsRegexp: extendsRegex,
		cache:         NewPermanentCache(),
	}
}

func buildExtendsRegexp() *regexp.Regexp {
	extendsRegex, err := regexp.Compile(`^{{[ \t]*extends[ \t]+"([^"]*)"[ \t]*}}[ \t]*\n`)
	if err != nil {
		panic(err)
	}
	return extendsRegex
}

func (e *Engine) Cache(cache Cache) {
	e.modifyMux.Lock()
	defer e.modifyMux.Unlock()

	if cache == nil {
		panic("cache must not be nil")
	}
	e.cache = cache
}

func (e *Engine) Funcs(funcMap FuncMap) *Engine {
	e.modifyMux.Lock()
	defer e.modifyMux.Unlock()

	e.funcMap = funcMap

	return e
}

func (e *Engine) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	tmpl, err := e.getTemplate(name)
	if err != nil {
		return fmt.Errorf("can't get template: %w", err)
	}

	return tmpl.Execute(wr, data)
}

func (e *Engine) getTemplate(filename string) (*template.Template, error) {
	tmpl := e.cache.get(filename)
	if tmpl == nil {
		return e.prepareTemplate(filename)
	}
	return tmpl, nil
}

func (e *Engine) prepareTemplate(filename string) (*template.Template, error) {
	e.modifyMux.Lock()
	defer e.modifyMux.Unlock()

	return e.getTemplateNoLock(filename)
}

func (e *Engine) getTemplateNoLock(filename string) (*template.Template, error) {
	tmpl := e.cache.get(filename)
	if tmpl != nil {
		return tmpl, nil
	}

	file, err := e.fs.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open template '%s' file: %w", filename, err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Printf("failed to close file after processing template %s: %s", filename, err)
		}
	}()

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read template '%s' contents: %w", filename, err)
	}

	if match := e.extendsRegexp.FindSubmatchIndex(contents); len(match) > 0 {
		parentName := string(contents[match[2]:match[3]])
		contents = contents[match[1]:]

		parentTmpl, err := e.getTemplateNoLock(parentName)
		if err != nil {
			return nil, fmt.Errorf("failed to get parent template '%s' for '%s': %w", parentName, filename, err)
		}

		tmpl, err = parentTmpl.Clone()
		if err != nil {
			return nil, fmt.Errorf("failed to clone parent template '%s' for '%s': %w", parentName, filename, err)
		}
	} else {
		tmpl = template.New("").Funcs(template.FuncMap(e.funcMap))
	}

	tmpl, err = tmpl.Parse(string(contents))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template '%s' contents: %w", filename, err)
	}

	e.cache.put(filename, tmpl)

	return tmpl, nil
}
