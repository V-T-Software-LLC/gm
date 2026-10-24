package ccpp

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := map[string]bool{
		"foo":     true,
		"foo_bar": true,
		"_foo":    true,
		"Foo2":    true,
		"foo-bar": false,
		"2foo":    false,
		"foo bar": false,
		"":        false,
	}
	for name, ok := range tests {
		if err := ValidateName(name); (err == nil) != ok {
			t.Errorf("ValidateName(%q) = %v, want ok=%v", name, err, ok)
		}
	}
}

var cmakeFiles = []string{
	".clang-format",
	".gitignore",
	"CMakeLists.txt",
	"cmake/DownloadGoogleTest.cmake",
	"cmake/helpers.cmake",
	"foo/CMakeLists.txt",
	"foo/copts/ConfigureCopts.cmake",
	"foo/copts/copts.py",
	"foo/copts/generate_copts.py",
}

var makeFiles = []string{
	".clang-format",
	".gitignore",
	"Makefile",
	"foo/Makefile",
}

func TestWriteFiles(t *testing.T) {
	tests := []struct {
		o        Options
		files    []string
		contains map[string]string
	}{
		{
			o:     Options{Build: "cmake", Language: "c", Kind: "lib", Compiler: "clang"},
			files: append([]string{"foo/foo.c", "foo/foo.h", "foo/foo_test.cc"}, cmakeFiles...),
			contains: map[string]string{
				"CMakeLists.txt":     "set(CMAKE_C_COMPILER clang)\nset(CMAKE_CXX_COMPILER clang++)",
				"foo/foo.h":          "#ifndef FOO_FOO_H_",
				"foo/CMakeLists.txt": "include(DownloadGoogleTest)",
			},
		},
		{
			o:     Options{Build: "cmake", Language: "c", Kind: "app", Compiler: "gcc"},
			files: append([]string{"foo/main.c"}, cmakeFiles...),
			contains: map[string]string{
				"CMakeLists.txt":     "set(CMAKE_C_COMPILER gcc)\nset(CMAKE_CXX_COMPILER g++)",
				"foo/CMakeLists.txt": "foo_add_executable(",
			},
		},
		{
			o:     Options{Build: "cmake", Language: "cpp", Kind: "lib", Compiler: "gcc"},
			files: append([]string{"foo/foo.cc", "foo/foo.h", "foo/foo_test.cc"}, cmakeFiles...),
			contains: map[string]string{
				"foo/foo.h":          "namespace foo {",
				"foo/CMakeLists.txt": "SRCS foo.cc",
			},
		},
		{
			o:     Options{Build: "cmake", Language: "cpp", Kind: "app", Compiler: "clang"},
			files: append([]string{"foo/main.cc"}, cmakeFiles...),
		},
		{
			o:     Options{Build: "make", Language: "c", Kind: "lib", Compiler: "gcc"},
			files: append([]string{"foo/foo.c", "foo/foo.h"}, makeFiles...),
			contains: map[string]string{
				"Makefile":     "$(MAKE) -C foo $@",
				"foo/Makefile": "CC=gcc\n",
			},
		},
		{
			o:     Options{Build: "make", Language: "c", Kind: "app", Compiler: "clang"},
			files: append([]string{"foo/main.c"}, makeFiles...),
			contains: map[string]string{
				"foo/Makefile": "BIN_NAME=foo\n",
			},
		},
	}

	for _, tt := range tests {
		tt.o.Name = "foo"
		t.Run(tt.o.Build+"-"+tt.o.Language+"-"+tt.o.Kind, func(t *testing.T) {
			chdir(t, t.TempDir())
			if err := writeFiles(tt.o); err != nil {
				t.Fatal(err)
			}

			got := listFiles(t, "foo")
			sort.Strings(tt.files)
			if !reflect.DeepEqual(got, tt.files) {
				t.Errorf("files = %v\nwant %v", got, tt.files)
			}
			for _, f := range got {
				b, _ := os.ReadFile(filepath.Join("foo", f))
				if strings.Contains(string(b), "{{") {
					t.Errorf("%s has an unexecuted template action", f)
				}
			}
			for f, s := range tt.contains {
				b, _ := os.ReadFile(filepath.Join("foo", f))
				if !strings.Contains(string(b), s) {
					t.Errorf("%s does not contain %q", f, s)
				}
			}
		})
	}
}

func TestGenerateCopts(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not installed")
	}
	chdir(t, t.TempDir())
	o := Options{Name: "foo", Build: "cmake", Language: "c", Kind: "lib", Compiler: "clang"}
	if err := writeFiles(o); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("python3", "-B", "generate_copts.py")
	cmd.Dir = filepath.Join("foo", "foo", "copts")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, out)
	}

	b, err := os.ReadFile(filepath.Join("foo", "foo", "copts", "GENERATED_fooCopts.cmake"))
	if err != nil {
		t.Fatal(err)
	}
	generated := string(b)
	for _, v := range []string{"foo_GCC_FLAGS", "foo_GCC_CXX_FLAGS", "foo_GCC_TEST_FLAGS", "foo_LLVM_FLAGS", "foo_LLVM_CXX_FLAGS", "foo_LLVM_TEST_FLAGS"} {
		if !strings.Contains(generated, "list(APPEND "+v+"\n") {
			t.Errorf("generated copts missing %s", v)
		}
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

func listFiles(t *testing.T, dir string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		files = append(files, filepath.ToSlash(rel))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(files)
	return files
}
