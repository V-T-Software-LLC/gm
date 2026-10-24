package ccpp

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/V-T-Software-LLC/gm/scaffold"
)

// cmake and make hold the files for the project root of each build system.
// src holds the sources and cmake-src the cmake files for each language and
// kind, both written to <name>/<name>.
//
//go:embed all:cmake cmake-src all:make src
var files embed.FS

// written to the root of every project
//
//go:embed .clang-format
var clangFormat []byte

type Options struct {
	Name     string
	Build    string // "cmake" or "make"
	Language string // "c" or "cpp", always "c" for make
	Kind     string // "lib" or "app"
	Compiler string // "clang" or "gcc"
}

func (o Options) CC() string {
	if o.Compiler == "gcc" {
		return "gcc"
	}
	return "clang"
}

func (o Options) CXX() string {
	if o.Compiler == "gcc" {
		return "g++"
	}
	return "clang++"
}

// the name is used in cmake, python, and C identifiers
var validName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func ValidateName(name string) error {
	if !validName.MatchString(name) {
		return fmt.Errorf("%q must contain only letters, numbers, and underscores, and not start with a number", name)
	}
	return nil
}

func Create(o Options) error {
	steps := []scaffold.Step{
		{Title: "Creating " + o.Name, Fn: func() error { return os.Mkdir(o.Name, 0o755) }},
		{Title: "Writing project files", Fn: func() error { return writeFiles(o) }},
	}
	if o.Build == "cmake" {
		steps = append(steps, scaffold.Step{Title: "Generating compile options", Fn: func() error {
			return scaffold.Run(filepath.Join(o.Name, o.Name, "copts"), "python3", "-B", "generate_copts.py")
		}})
	}
	if err := scaffold.RunSteps(steps); err != nil {
		return err
	}

	fmt.Printf("\nCreated %s. To build:\n\n", o.Name)
	fmt.Printf("  cd %s\n", o.Name)
	switch {
	case o.Build == "make" && o.Kind == "lib":
		fmt.Println("  make")
	case o.Build == "make":
		fmt.Println("  make")
		fmt.Printf("  ./%s/%s\n", o.Name, o.Name)
	case o.Kind == "lib":
		fmt.Println("  cmake -B build")
		fmt.Println("  cmake --build build")
		fmt.Println("  ctest --test-dir build")
	default:
		fmt.Println("  cmake -B build")
		fmt.Println("  cmake --build build")
		fmt.Printf("  ./build/%s/%s\n", o.Name, o.Name)
	}
	return nil
}

func writeFiles(o Options) error {
	dirs := map[string]string{
		o.Build:                              o.Name,
		path.Join("src", o.Language, o.Kind): filepath.Join(o.Name, o.Name),
	}
	if o.Build == "cmake" {
		dirs[path.Join("cmake-src", o.Language, o.Kind)] = filepath.Join(o.Name, o.Name)
	}

	for src, dst := range dirs {
		sub, err := fs.Sub(files, src)
		if err != nil {
			return err
		}
		if err := scaffold.Write(sub, dst, o); err != nil {
			return err
		}
	}
	return os.WriteFile(filepath.Join(o.Name, ".clang-format"), clangFormat, 0o644)
}
