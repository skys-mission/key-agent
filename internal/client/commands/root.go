// Package commands provides CLI commands for keyctl.
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version string
	commit  string
	date    string
)

var (
	apiAddr   string
	token     string
	tokenFile string
)

var rootCmd = &cobra.Command{
	Use:   "keyctl",
	Short: "key-agent CLI client",
	Long:  `keyctl is the command-line client for key-agent daemon.`,
}

// Execute runs the root command.
func Execute(v, c, d string) error {
	version, commit, date = v, c, d
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&apiAddr, "addr", "a", "http://127.0.0.1:8080", "API server address")
	rootCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "API token")
	rootCmd.PersistentFlags().StringVarP(&tokenFile, "token-file", "f", "", "Path to token file (default: ~/.skys-mission/key-agent/data/token)")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(kvCmd)
	rootCmd.AddCommand(secretCmd)
	rootCmd.AddCommand(tokenCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("keyctl %s (commit: %s, built: %s)\n", version, commit, date)
	},
}

var kvCmd = &cobra.Command{
	Use:   "kv",
	Short: "KV operations",
}

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Secret operations",
}

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Token operations",
}
