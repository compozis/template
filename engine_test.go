package template_test

import (
	"bytes"
	"github.com/compozis/template"
	"gotest.tools/assert"
	"io/ioutil"
	"path"
	"runtime"
	"testing"
)

func TestEngine_ExecuteTemplate(t *testing.T) {
	tests := []struct {
		filename string
	}{
		{"article_page"},
		{"plain_page"},
		{"search"},
	}

	engine := template.NewEngine(template.Dir(resourcesDir()))
	engine.Partials("component/echo.gohtml")

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			var buf bytes.Buffer
			err := engine.ExecuteTemplate(&buf, test.filename+".gohtml", nil)
			assert.NilError(t, err)
			assert.DeepEqual(t, buf.Bytes(), readFile(test.filename+".out"))
		})
	}
}

type nopWriter struct{}

func (n2 nopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

type countAccessFileSystem struct {
	delegate template.FileSystem
	count    int
}

func (c *countAccessFileSystem) Open(name string) (template.File, error) {
	c.count += 1
	return c.delegate.Open(name)
}

func TestEngine_Cache(t *testing.T) {
	tests := []struct {
		name               string
		cache              template.Cache
		templatesToExecute []string
		count              int
	}{
		{
			"NoCache_Base",
			template.NoCache{},
			[]string{"article_page.gohtml", "plain_page.gohtml", "search.gohtml"},
			10,
		},
		{
			"NoCache_ReExecute",
			template.NoCache{},
			[]string{"article_page.gohtml", "plain_page.gohtml", "search.gohtml", "search.gohtml"},
			14,
		},
		{
			"PermanentCache_Base",
			template.NewPermanentCache(),
			[]string{"article_page.gohtml", "plain_page.gohtml", "search.gohtml"},
			6,
		},
		{
			"PermanentCache_ReExecute",
			template.NewPermanentCache(),
			[]string{"article_page.gohtml", "plain_page.gohtml", "search.gohtml", "search.gohtml"},
			6,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := &countAccessFileSystem{
				delegate: template.Dir(resourcesDir()),
				count:    0,
			}

			engine := template.NewEngine(fs)
			engine.Partials("component/echo.gohtml")
			engine.Cache(test.cache)

			for _, templateName := range test.templatesToExecute {
				var wr nopWriter
				err := engine.ExecuteTemplate(wr, templateName, nil)
				assert.NilError(t, err)
			}

			assert.Equal(t, fs.count, test.count)
		})
	}
}

func TestEngine_Funcs(t *testing.T) {
	engine := template.NewEngine(template.Dir(resourcesDir()))
	engine.Funcs(template.FuncMap{
		"plus":  func(a, b int) int { return a + b },
		"minus": func(a, b int) int { return a - b },
	})

	var buf bytes.Buffer
	err := engine.ExecuteTemplate(&buf, "funcs.gohtml", nil)
	assert.NilError(t, err)
	assert.DeepEqual(t, buf.Bytes(), readFile("funcs.out"))
}

func readFile(name string) []byte {
	result, err := ioutil.ReadFile(path.Join(resourcesDir(), name))
	if err != nil {
		panic(err)
	}
	return result
}

func resourcesDir() string {
	_, filePath, _, _ := runtime.Caller(0)

	return path.Join(path.Dir(filePath), "test", "data")
}
