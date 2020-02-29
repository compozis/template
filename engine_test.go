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

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			var buf bytes.Buffer
			err := engine.ExecuteTemplate(&buf, test.filename+".gohtml", nil)
			assert.NilError(t, err)
			assert.DeepEqual(t, buf.Bytes(), readFile(test.filename+".out"))
		})
	}
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
