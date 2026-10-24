package webapp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFiles(t *testing.T) {
	for _, air := range []bool{true, false} {
		name := "without air"
		if air {
			name = "with air"
		}
		t.Run(name, func(t *testing.T) {
			chdir(t, t.TempDir())
			o := Options{Name: "demo", Module: "example.com/demo", PackageManager: "pnpm", Air: air}
			if err := os.Mkdir(o.Name, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := writeFiles(o); err != nil {
				t.Fatal(err)
			}

			contains := map[string]string{
				"cmd/demo/main.go":          `"example.com/demo/internal/server"`,
				"internal/server/server.go": `"example.com/demo/web/views"`,
				"Makefile":                  "PM := pnpm",
				"web/views/layout.templ":    "if Dev {",
				"web/app/vite.config.js":    "outDir: '../static'",
				".gitignore":                "bin/",
				"sqlc.yaml":                 `out: "internal/db"`,
			}
			if air {
				contains[".air.toml"] = `cmd = "go build -o ./tmp/main ./cmd/demo"`
				contains["Makefile"] = "dev-go:\n\tair\n"
			} else {
				contains["Makefile"] = "go run ./cmd/$(NAME) -dev"
			}

			for f, s := range contains {
				b, err := os.ReadFile(filepath.Join(o.Name, f))
				if err != nil {
					t.Error(err)
					continue
				}
				if !strings.Contains(string(b), s) {
					t.Errorf("%s does not contain %q", f, s)
				}
			}

			_, err := os.Stat(filepath.Join(o.Name, ".air.toml"))
			if exists := err == nil; exists != air {
				t.Errorf(".air.toml exists = %v, want %v", exists, air)
			}
		})
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(wd) })
}
