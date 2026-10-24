/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gm",
	Short: "Scaffold new projects",
	Long: `gm scaffolds new projects. Run gm create and answer the prompts.

Project types:
  web app    go server with echo, templ, and sqlc, and a vite svelte app
             with tailwind for dynamic components
  low level  C or C++ library or application built with CMake, or C with
             a basic Makefile`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
