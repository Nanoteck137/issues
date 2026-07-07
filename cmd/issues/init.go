package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nanoteck137/issues/config"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

var initOutput string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := initOutput
		if configPath == "" {
			configDir, err := config.DefaultConfigDir()
			if err != nil {
				return fmt.Errorf("getting config directory: %w", err)
			}
			configPath = filepath.Join(configDir, "config.toml")
		}

		if _, err := os.Stat(configPath); err == nil {
			fmt.Print("Config already exists. Overwrite? [y/N]: ")
			var answer string
			fmt.Scanln(&answer)
			if strings.ToLower(answer) != "y" {
				fmt.Println("Aborting.")
				return nil
			}
		}

		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return fmt.Errorf("creating config directory: %w", err)
		}

		scanner := bufio.NewScanner(os.Stdin)

		server := prompt(scanner, "Server", "https://forgejo.nanoteck137.net")
		fmt.Print("Token: ")
		scanner.Scan()
		token := strings.TrimSpace(scanner.Text())

		type configFile struct {
			Server string `toml:"server"`
			Token  string `toml:"token"`
		}

		cfg := configFile{
			Server: server,
			Token:  token,
		}

		data, err := toml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("marshaling config: %w", err)
		}

		if err := os.WriteFile(configPath, data, 0600); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}

		fmt.Printf("Config written to %s\n", configPath)
		return nil
	},
}

func prompt(scanner *bufio.Scanner, label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("%s: ", label)
	}
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return defaultVal
	}
	return input
}

func init() {
	initCmd.Flags().StringVarP(&initOutput, "output", "o", "", "Output path for the config file")

	rootCmd.AddCommand(initCmd)
}
