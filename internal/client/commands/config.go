package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// CLIConfig holds keyctl configuration.
type CLIConfig struct {
	Addr      string `yaml:"addr,omitempty"`
	Token     string `yaml:"token,omitempty"`
	TokenFile string `yaml:"token_file,omitempty"`
}

// ConfigPath returns the default keyctl config file path.
func ConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".skys-mission", "key-agent", "keyctl.yaml")
}

// LoadConfig loads CLI configuration from file.
func LoadConfig() (*CLIConfig, error) {
	cfg := &CLIConfig{}

	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save saves CLI configuration to file.
func (c *CLIConfig) Save() error {
	path := ConfigPath()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// GetEffectiveConfig returns the effective configuration.
// Priority: command line > config file > environment variable
func GetEffectiveConfig() (*CLIConfig, error) {
	// Start with empty config
	effective := &CLIConfig{}

	// 1. Load from config file (lowest priority)
	fileCfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	*effective = *fileCfg

	// 2. Override with environment variables
	if envAddr := os.Getenv("KEY_AGENT_ADDR"); envAddr != "" {
		effective.Addr = envAddr
	}
	if envToken := os.Getenv("KEY_AGENT_TOKEN"); envToken != "" {
		effective.Token = envToken
	}

	// 3. Override with command line flags (highest priority)
	if apiAddr != "" {
		effective.Addr = apiAddr
	}
	if token != "" {
		effective.Token = token
	}
	if tokenFile != "" {
		effective.TokenFile = tokenFile
	}

	return effective, nil
}

func init() {
	// Add config commands
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage keyctl configuration",
	Long:  `View and modify keyctl configuration settings.`,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := LoadConfig()
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error loading config: %v\n", err)
			os.Exit(1)
		}

		key := args[0]
		switch key {
		case "addr":
			fmt.Println(cfg.Addr)
		case "token":
			fmt.Println(cfg.Token)
		case "token_file":
			fmt.Println(cfg.TokenFile)
		default:
			fmt.Fprintf(cmd.ErrOrStderr(), "Unknown config key: %s\n", key)
			os.Exit(1)
		}
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := LoadConfig()
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error loading config: %v\n", err)
			os.Exit(1)
		}

		key, value := args[0], args[1]
		switch key {
		case "addr":
			cfg.Addr = value
		case "token":
			cfg.Token = value
		case "token_file":
			cfg.TokenFile = value
		default:
			fmt.Fprintf(cmd.ErrOrStderr(), "Unknown config key: %s\n", key)
			os.Exit(1)
		}

		if err := cfg.Save(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Set %s = %s\n", key, value)
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := LoadConfig()
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error loading config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Config file: %s\n", ConfigPath())
		fmt.Printf("  addr:       %s\n", cfg.Addr)
		if cfg.Token != "" {
			fmt.Printf("  token:      %s...\n", cfg.Token[:min(len(cfg.Token), 8)])
		} else {
			fmt.Printf("  token:      (not set)\n")
		}
		fmt.Printf("  token_file: %s\n", cfg.TokenFile)
	},
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}