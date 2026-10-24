package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/V-T-Software-LLC/gm/webapp"
)

type ProjectType string

const (
	WebApp   ProjectType = "web app"
	CLI      ProjectType = "cli"
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
						huh.NewOption("CLI", CLI),
						huh.NewOption("Low level", LowLevel),
					).
					Value(&p.Type),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		switch p.Type {
		case WebApp:
			return createWebApp(p.Name)
		default:
			return fmt.Errorf("%s projects are not supported yet", p.Type)
		}
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
	)

	if err := form.Run(); err != nil {
		return err
	}

	return webapp.Create(o)
}

func init() {
	rootCmd.AddCommand(createCmd)
}
