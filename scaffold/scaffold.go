package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/charmbracelet/lipgloss"
)

var (
	check = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✓")
	cross = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("✗")
	Faint = lipgloss.NewStyle().Faint(true)
)

type Step struct {
	Title string
	Fn    func() error
}

// RunSteps runs each step, printing its title while it runs and a check or
// cross when it finishes.
func RunSteps(steps []Step) error {
	fmt.Println()
	for _, s := range steps {
		fmt.Print(Faint.Render("  " + s.Title + "..."))
		if err := s.Fn(); err != nil {
			fmt.Printf("\r\033[K%s %s\n\n", cross, s.Title)
			return err
		}
		fmt.Printf("\r\033[K%s %s\n", check, s.Title)
	}
	return nil
}

// Write copies fsys into dir. Paths are executed as templates with data.
// Files ending in .tmpl are executed as templates and written without the suffix.
func Write(fsys fs.FS, dir string, data any) error {
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		dst, err := Execute(path, data)
		if err != nil {
			return err
		}
		dst = filepath.Join(dir, dst)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}

		b, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".tmpl") {
			return os.WriteFile(dst, b, 0o644)
		}

		s, err := Execute(string(b), data)
		if err != nil {
			return err
		}
		return os.WriteFile(strings.TrimSuffix(dst, ".tmpl"), []byte(s), 0o644)
	})
}

func Execute(text string, data any) (string, error) {
	t, err := template.New("").Funcs(template.FuncMap{"upper": strings.ToUpper}).Parse(text)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if err := t.Execute(&b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}

// Run runs a command in dir. Its output is only shown if it fails.
func Run(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		os.Stderr.Write(out)
		return fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}
