package cmd

import (
	"errors"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	ccpp "github.com/V-T-Software-LLC/gm/c-cpp"
	"github.com/V-T-Software-LLC/gm/webapp"
)

type ProjectType string

const (
	WebApp   ProjectType = "web app"
	LowLevel ProjectType = "low level"
)

type Project struct {
	Name string
	Type ProjectType
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project",
	RunE: func(cmd *cobra.Command, args []string) error {
		var p Project

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Project name").
					Value(&p.Name).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return errors.New("name is required")
						}
						return nil
					}),
			),
			huh.NewGroup(
				huh.NewSelect[ProjectType]().
					Title("Project type").
					Options(
						huh.NewOption("Web app", WebApp),
						huh.NewOption("Low level", LowLevel),
					).
					Value(&p.Type),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		if p.Type == WebApp {
			return createWebApp(p.Name)
		}
		return createLowLevel(p.Name)
	},
}

func createWebApp(name string) error {
	o := webapp.Options{Name: name}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Go module name").
				Value(&o.Module).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("module name is required")
					}
					return nil
				}),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("JavaScript package manager").
				Options(huh.NewOptions("npm", "pnpm", "bun")...).
				Value(&o.PackageManager),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Language").
				Options(
					huh.NewOption("JavaScript", "js"),
					huh.NewOption("TypeScript", "ts"),
				).
				Value(&o.Lang),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	return webapp.Create(o)
}

func createLowLevel(name string) error {
	if err := ccpp.ValidateName(name); err != nil {
		return err
	}

	o := ccpp.Options{Name: name, Language: "c"}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Build system").
				Options(
					huh.NewOption("CMake", "cmake"),
					huh.NewOption("Makefile (C only)", "make"),
				).
				Value(&o.Build),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Language").
				Options(
					huh.NewOption("C", "c"),
					huh.NewOption("C++", "cpp"),
				).
				Value(&o.Language),
		).WithHideFunc(func() bool { return o.Build == "make" }),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Library or application").
				Options(
					huh.NewOption("Library", "lib"),
					huh.NewOption("Application", "app"),
				).
				Value(&o.Kind),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Compiler").
				Options(
					huh.NewOption("Clang", "clang"),
					huh.NewOption("GCC", "gcc"),
				).
				Value(&o.Compiler),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}
	if o.Build == "make" {
		o.Language = "c"
	}

	return ccpp.Create(o)
}

func init() {
	rootCmd.AddCommand(createCmd)
}
