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
			o := Options{Name: "demo", Module: "example.com/demo", PackageManager: "pnpm", Lang: "ts", Air: air}
			if err := os.Mkdir(o.Name, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := writeFiles(o); err != nil {
				t.Fatal(err)
			}

			contains := map[string]string{
				"cmd/demo/main.go":                              `"example.com/demo/internal/server"`,
				"internal/server/server.go":                     `"example.com/demo/web/views"`,
				"Makefile":                                      "PM := pnpm",
				"web/views/layout.templ":                        `viteURL + "/src/main.ts"`,
				"web/app/vite.config.ts":                        "input: ['src/main.ts', 'src/app.css']",
				"web/app/src/main.ts":                           "mount(App, { target })",
				"web/app/src/App.svelte":                        `<script lang="ts">`,
				".gitignore":                                    "bin/",
				"sqlc.yaml":                                     `schema: "internal/db/migrations"`,
				"internal/server/auth.go":                       `"example.com/demo/internal/db"`,
				"internal/server/session.go":                    "session:",
				"internal/db/migrations/00001_create_users.sql": "-- +goose Up",
				"web/views/dashboard.templ":                     `import "example.com/demo/internal/db"`,
				".env.example":                                  "localhost:5432/demo",
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
