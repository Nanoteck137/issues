package main

import (
	"os"

	"github.com/kr/pretty"
	"github.com/nanoteck137/issues"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     issues.AppName,
	Version: issues.Version,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func init() {
	rootCmd.SetVersionTemplate(issues.VersionTemplate(issues.AppName))
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		pretty.Println(err)
		os.Exit(1)
	}
}
