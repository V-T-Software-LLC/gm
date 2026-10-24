package webapp

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/charmbracelet/lipgloss"
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

var (
	check = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✓")
	cross = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("✗")
	faint = lipgloss.NewStyle().Faint(true)
)

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

	steps := []struct {
		title string
		fn    func() error
	}{
		{"Creating " + o.Name, func() error { return os.Mkdir(o.Name, 0o755) }},
		{"Initializing go module", func() error { return run(o.Name, "go", "mod", "init", o.Module) }},
		{"Creating vite svelte app", func() error { return createVite(o.Name, pm, create) }},
		{"Writing project files", func() error { return writeFiles(o) }},
		{"Installing tailwind", func() error { return run(web, pm, add...) }},
		{"Installing echo and templ", func() error {
			return run(o.Name, "go", "get", "github.com/labstack/echo/v5", "github.com/a-h/templ")
		}},
		{"Generating templ files", func() error { return run(o.Name, "templ", "generate") }},
		{"Tidying go module", func() error { return run(o.Name, "go", "mod", "tidy") }},
	}

	fmt.Println()
	for _, s := range steps {
		fmt.Print(faint.Render("  " + s.title + "..."))
		if err := s.fn(); err != nil {
			fmt.Printf("\r\033[K%s %s\n\n", cross, s.title)
			return err
		}
		fmt.Printf("\r\033[K%s %s\n", check, s.title)
	}

	fmt.Printf("\nCreated %s. To start developing:\n\n", o.Name)
	fmt.Printf("  cd %s\n", o.Name)
	fmt.Println("  make dev")
	fmt.Println("\nThen open http://localhost:8080")
	if !o.Air {
		fmt.Println(faint.Render("\nInstall air for live reload: go install github.com/air-verse/air@latest"))
	}
	return nil
}

func createVite(dir, pm string, args []string) error {
	if err := run(dir, pm, args...); err != nil {
		return err
	}
	for _, p := range viteLeftovers {
		if err := os.RemoveAll(filepath.Join(dir, p)); err != nil {
			return err
		}
	}
	return nil
}

// writeFiles copies the embedded files into the project. Paths are executed as
// templates with the options. Files ending in .tmpl are executed as templates
// and written without the suffix.
func writeFiles(o Options) error {
	if o.Air {
		s, err := execute(airConfig, o)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(o.Name, ".air.toml"), []byte(s), 0o644); err != nil {
			return err
		}
	}

	return fs.WalkDir(files, "files", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		dst, err := execute(strings.TrimPrefix(path, "files"), o)
		if err != nil {
			return err
		}
		dst = filepath.Join(o.Name, dst)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}

		data, err := files.ReadFile(path)
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".tmpl") {
			return os.WriteFile(dst, data, 0o644)
		}

		s, err := execute(string(data), o)
		if err != nil {
			return err
		}
		return os.WriteFile(strings.TrimSuffix(dst, ".tmpl"), []byte(s), 0o644)
	})
}

func execute(text string, o Options) (string, error) {
	t, err := template.New("").Parse(text)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if err := t.Execute(&b, o); err != nil {
		return "", err
	}
	return b.String(), nil
}

// run runs a command in dir. Its output is only shown if it fails.
func run(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		os.Stderr.Write(out)
		return fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}
