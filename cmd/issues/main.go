package main

import (
	"os"

	"github.com/kr/pretty"
	"github.com/nanoteck137/issues"
	"github.com/nanoteck137/issues/config"
	"github.com/spf13/cobra"
)

var (
	serverFlag  string
	ownerFlag   string
	tokenFlag   string
	jsonFlag    bool
	verboseFlag bool
	configFlag  string

	gitInfo = issues.DetectFromGit()
)

type ResolvedConfig struct {
	Server  string
	Owner   string
	Token   string
	Repo    string
	JSON    bool
	Verbose bool
}

var rootCmd = &cobra.Command{
	Use:     issues.AppName,
	Version: issues.Version,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func init() {
	rootCmd.SetVersionTemplate(issues.VersionTemplate(issues.AppName))

	rootCmd.PersistentFlags().StringVar(&configFlag, "config", "", "Config file path")
	rootCmd.PersistentFlags().StringVar(&serverFlag, "server", "", "Forgejo server URL")
	rootCmd.PersistentFlags().StringVar(&ownerFlag, "owner", "", "Repository owner")
	rootCmd.PersistentFlags().StringVar(&tokenFlag, "token", "", "Authentication token")
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Verbose output")
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		pretty.Println(err)
		os.Exit(1)
	}
}

func loadConfig() *config.Config {
	cfg, err := config.Load(configFlag)
	if err != nil {
		return &config.Config{
			Server: "https://forgejo.nanoteck137.net",
		}
	}
	return cfg
}

func resolveConfig(cmd *cobra.Command) ResolvedConfig {
	cfg := loadConfig()

	server := serverFlag
	if !cmd.Flags().Changed("server") {
		if cfg.Server != "" {
			server = cfg.Server
		} else if gitInfo.Server != "" {
			server = gitInfo.Server
		} else {
			server = "https://forgejo.nanoteck137.net"
		}
	}

	owner := ownerFlag
	if !cmd.Flags().Changed("owner") && gitInfo.Owner != "" {
		owner = gitInfo.Owner
	}

	token := tokenFlag
	if !cmd.Flags().Changed("token") && cfg.Token != "" {
		token = cfg.Token
	}

	repo := ""
	if cmd.Flags().Lookup("repo") != nil {
		repo = cmd.Flags().Lookup("repo").Value.String()
		if !cmd.Flags().Changed("repo") && gitInfo.Repo != "" {
			repo = gitInfo.Repo
		}
	}

	return ResolvedConfig{
		Server:  server,
		Owner:   owner,
		Token:   token,
		Repo:    repo,
		JSON:    jsonFlag,
		Verbose: verboseFlag,
	}
}
