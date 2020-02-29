package template

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"regexp"
)

type Engine struct {
	fs FileSystem

	extendsRegexp *regexp.Regexp
}

func NewEngine(fs FileSystem) *Engine {
	extendsRegex := buildExtendsRegexp()

	return &Engine{
		fs: fs,

		extendsRegexp: extendsRegex,
	}
}

func buildExtendsRegexp() *regexp.Regexp {
	extendsRegex, err := regexp.Compile(`^{{[ \t]*extends[ \t]+"([^"]*)"[ \t]*}}[ \t]*\n`)
	if err != nil {
		panic(err)
	}
	return extendsRegex
}

func (e *Engine) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	// TODO: cache support
	tmpl, err := e.getTemplate(name)
	if err != nil {
		return fmt.Errorf("can't get template: %w", err)
	}

	return tmpl.Execute(wr, data)
}

func (e *Engine) getTemplate(filename string) (*template.Template, error) {
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

	var tmpl *template.Template

	if match := e.extendsRegexp.FindSubmatchIndex(contents); len(match) > 0 {
		parentName := string(contents[match[2]:match[3]])
		contents = contents[match[1]:]

		parentTmpl, err := e.getTemplate(parentName)
		if err != nil {
			return nil, fmt.Errorf("failed to get parent template '%s' for '%s': %w", parentName, filename, err)
		}

		tmpl, err = parentTmpl.Clone()
		if err != nil {
			return nil, fmt.Errorf("failed to clone parent template '%s' for '%s': %w", parentName, filename, err)
		}
	} else {
		tmpl = template.New("")
	}

	tmpl, err = tmpl.Parse(string(contents))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template '%s' contents: %w", filename, err)
	}

	return tmpl, nil
}
