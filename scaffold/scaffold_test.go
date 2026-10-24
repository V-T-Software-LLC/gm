package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestWrite(t *testing.T) {
	fsys := fstest.MapFS{
		"{{.Name}}/main.c.tmpl": {Data: []byte("// {{upper .Name}} {{.Name}}\n")},
		"plain.txt":             {Data: []byte("{{.Name}} is not executed\n")},
	}
	dir := t.TempDir()

	if err := Write(fsys, dir, struct{ Name string }{"foo"}); err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"foo/main.c": "// FOO foo\n",
		"plain.txt":  "{{.Name}} is not executed\n",
	}
	for path, content := range want {
		b, err := os.ReadFile(filepath.Join(dir, path))
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != content {
			t.Errorf("%s = %q, want %q", path, b, content)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "foo", "main.c.tmpl")); err == nil {
		t.Error("main.c.tmpl should be written without the .tmpl suffix")
	}
}

func TestWriteBadTemplate(t *testing.T) {
	fsys := fstest.MapFS{"bad.tmpl": {Data: []byte("${{{.Name}}}")}}
	if err := Write(fsys, t.TempDir(), struct{ Name string }{"foo"}); err == nil {
		t.Error("expected a parse error")
	}
}

func TestRun(t *testing.T) {
	if err := Run(t.TempDir(), "go", "version"); err != nil {
		t.Fatal(err)
	}

	err := Run(t.TempDir(), "go", "notacommand")
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "go notacommand") {
		t.Errorf("error %q should name the command", err)
	}
}
