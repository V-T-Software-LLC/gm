package webapp

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/V-T-Software-LLC/gm/scaffold"
)

//go:embed all:files
var files embed.FS

// only written when air is installed
//
//go:embed air.toml.tmpl
var airConfig string

type Options struct {
	Name           string
	Module         string
	PackageManager string
	Air            bool
}

// vite template files replaced by our own build setup
var viteLeftovers = []string{
	"web/app/index.html",
	"web/app/public",
	"web/app/src/assets",
	"web/app/src/lib",
}

func Create(o Options) error {
	_, err := exec.LookPath("air")
	o.Air = err == nil

	pm := o.PackageManager
	web := filepath.Join(o.Name, "web", "app")
	create := []string{"create", "vite", "web/app", "--template", "svelte", "--no-interactive"}
	add := []string{"add", "-D", "tailwindcss", "@tailwindcss/vite"}
	if pm == "npm" {
		create = []string{"create", "vite@latest", "web/app", "--", "--template", "svelte", "--no-interactive"}
		add[0] = "install"
	}

	err = scaffold.RunSteps([]scaffold.Step{
		{Title: "Creating " + o.Name, Fn: func() error { return os.Mkdir(o.Name, 0o755) }},
		{Title: "Initializing go module", Fn: func() error { return scaffold.Run(o.Name, "go", "mod", "init", o.Module) }},
		{Title: "Creating vite svelte app", Fn: func() error { return createVite(o.Name, pm, create) }},
		{Title: "Writing project files", Fn: func() error { return writeFiles(o) }},
		{Title: "Installing tailwind", Fn: func() error { return scaffold.Run(web, pm, add...) }},
		{Title: "Installing echo and templ", Fn: func() error {
			return scaffold.Run(o.Name, "go", "get", "github.com/labstack/echo/v5", "github.com/a-h/templ")
		}},
		{Title: "Generating templ files", Fn: func() error { return scaffold.Run(o.Name, "templ", "generate") }},
		{Title: "Tidying go module", Fn: func() error { return scaffold.Run(o.Name, "go", "mod", "tidy") }},
	})
	if err != nil {
		return err
	}

	fmt.Printf("\nCreated %s. To start developing:\n\n", o.Name)
	fmt.Printf("  cd %s\n", o.Name)
	fmt.Println("  make dev")
	fmt.Println("\nThen open http://localhost:8080")
	if !o.Air {
		fmt.Println(scaffold.Faint.Render("\nInstall air for live reload: go install github.com/air-verse/air@latest"))
	}
	return nil
}

func createVite(dir, pm string, args []string) error {
	if err := scaffold.Run(dir, pm, args...); err != nil {
		return err
	}
	for _, p := range viteLeftovers {
		if err := os.RemoveAll(filepath.Join(dir, p)); err != nil {
			return err
		}
	}
	return nil
}

func writeFiles(o Options) error {
	if o.Air {
		s, err := scaffold.Execute(airConfig, o)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(o.Name, ".air.toml"), []byte(s), 0o644); err != nil {
			return err
		}
	}

	sub, err := fs.Sub(files, "files")
	if err != nil {
		return err
	}
	return scaffold.Write(sub, o.Name, o)
}
