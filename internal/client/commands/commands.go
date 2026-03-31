package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/skys-mission/key-agent/keysdk"
	"github.com/spf13/cobra"
)

func init() {
	// KV commands
	kvGetCmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a KV value",
		Args:  cobra.ExactArgs(1),
		Run:   kvGet,
	}
	kvGetCmd.Flags().Bool("raw", false, "Output raw value only")

	kvSetCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a KV value",
		Args:  cobra.ExactArgs(2),
		Run:   kvSet,
	}

	kvDeleteCmd := &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete a KV entry",
		Args:  cobra.ExactArgs(1),
		Run:   kvDelete,
	}

	kvListCmd := &cobra.Command{
		Use:   "list",
		Short: "List KV keys",
		Args:  cobra.NoArgs,
		Run:   kvList,
	}
	kvListCmd.Flags().String("prefix", "", "Filter keys by prefix")

	kvCmd.AddCommand(kvGetCmd)
	kvCmd.AddCommand(kvSetCmd)
	kvCmd.AddCommand(kvDeleteCmd)
	kvCmd.AddCommand(kvListCmd)

	// Secret commands
	secretGetCmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a secret value",
		Args:  cobra.ExactArgs(1),
		Run:   secretGet,
	}
	secretGetCmd.Flags().Bool("raw", false, "Output raw value only")

	secretSetCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a secret value",
		Args:  cobra.ExactArgs(2),
		Run:   secretSet,
	}
	secretSetCmd.Flags().String("type", "password", "Secret type (password, api_key, certificate, private_key, token, other)")

	secretDeleteCmd := &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete a secret entry",
		Args:  cobra.ExactArgs(1),
		Run:   secretDelete,
	}

	secretListCmd := &cobra.Command{
		Use:   "list",
		Short: "List secret keys",
		Args:  cobra.NoArgs,
		Run:   secretList,
	}
	secretListCmd.Flags().String("prefix", "", "Filter keys by prefix")

	secretCmd.AddCommand(secretGetCmd)
	secretCmd.AddCommand(secretSetCmd)
	secretCmd.AddCommand(secretDeleteCmd)
	secretCmd.AddCommand(secretListCmd)

	// Token commands
	tokenCreateCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new token",
		Args:  cobra.ExactArgs(1),
		Run:   tokenCreate,
	}
	tokenCreateCmd.Flags().String("type", "client", "Token type (client, mcp)")
	tokenCreateCmd.Flags().String("expires-in", "", "Expiration duration (e.g., 24h, 30d)")

	tokenSaveCmd := &cobra.Command{
		Use:   "save <token>",
		Short: "Save token to default location (~/.key-agent/token)",
		Args:  cobra.ExactArgs(1),
		Run:   tokenSave,
	}

	tokenCmd.AddCommand(tokenCreateCmd)
	tokenCmd.AddCommand(tokenSaveCmd)
}

// KV command implementations

func kvGet(cmd *cobra.Command, args []string) {
	key := args[0]
	c := keysdk.NewClient(&keysdk.Config{
		BaseURL: apiAddr,
		Token:   getToken(),
	})

	entry, err := c.GetKV(key)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		os.Exit(1)
		return
	}

	if raw, _ := cmd.Flags().GetBool("raw"); raw {
		fmt.Println(entry.Value)
		return
	}

	data, _ := json.MarshalIndent(entry, "", "  ")
	fmt.Println(string(data))
}

func kvSet(cmd *cobra.Command, args []string) {
	key := args[0]
	value := args[1]
	c := keysdk.NewClient(&keysdk.Config{
		BaseURL: apiAddr,
		Token:   getToken(),
	})

	entry, err := c.SetKV(key, &keysdk.SetKVOptions{
		Value: value,
	})
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		os.Exit(1)
		return
	}

	data, _ := json.MarshalIndent(entry, "", "  ")
	fmt.Println(string(data))
}

func kvDelete(cmd *cobra.Command, args []string) {
	key := args[0]
	c := keysdk.NewClient(&keysdk.Config{
		BaseURL: apiAddr,
		Token:   getToken(),
	})

	if err := c.DeleteKV(key); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		os.Exit(1)
		return
	}

	fmt.Printf("Deleted key: %s\n", key)
}

func kvList(cmd *cobra.Command, args []string) {
	prefix, _ := cmd.Flags().GetString("prefix")
	c := keysdk.NewClient(&keysdk.Config{
		BaseURL: apiAddr,
		Token:   getToken(),
	})

	keys, err := c.ListKV(prefix)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		os.Exit(1)
		return
	}

	for _, key := range keys {
		fmt.Println(key)
	}
}

// Secret command implementations

func secretGet(cmd *cobra.Command, args []string) {
	key := args[0]
	c := keysdk.NewClient(&keysdk.Config{
		BaseURL: apiAddr,
		Token:   getToken(),
	})

	entry, err := c.GetSecret(key)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		os.Exit(1)
		return
	}

	if raw, _ := cmd.Flags().GetBool("raw"); raw {
		fmt.Println(entry.Value)
		return
	}

	data, _ := json.MarshalIndent(entry, "", "  ")
	fmt.Println(string(data))
}

func secretSet(cmd *cobra.Command, args []string) {
	key := args[0]
	value := args[1]
	secretType, _ := cmd.Flags().GetString("type")
	c := keysdk.NewClient(&keysdk.Config{
		BaseURL: apiAddr,
		Token:   getToken(),
	})

	if secretType == "" {
		secretType = "other"
	}

	entry, err := c.SetSecret(key, &keysdk.SetSecretOptions{
		Value: value,
		Type:  keysdk.SecretType(secretType),
	})
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		os.Exit(1)
		return
	}

	data, _ := json.MarshalIndent(entry, "", "  ")
	fmt.Println(string(data))
}

func secretDelete(cmd *cobra.Command, args []string) {
	key := args[0]
	c := keysdk.NewClient(&keysdk.Config{
		BaseURL: apiAddr,
		Token:   getToken(),
	})

	if err := c.DeleteSecret(key); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		os.Exit(1)
		return
	}

	fmt.Printf("Deleted secret: %s\n", key)
}

func secretList(cmd *cobra.Command, args []string) {
	prefix, _ := cmd.Flags().GetString("prefix")
	c := keysdk.NewClient(&keysdk.Config{
		BaseURL: apiAddr,
		Token:   getToken(),
	})

	keys, err := c.ListSecret(prefix)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		os.Exit(1)
		return
	}

	for _, key := range keys {
		fmt.Println(key)
	}
}

// Token command implementations

func tokenCreate(cmd *cobra.Command, args []string) {
	name := args[0]
	tokenType, _ := cmd.Flags().GetString("type")
	expiresIn, _ := cmd.Flags().GetString("expires-in")
	c := keysdk.NewClient(&keysdk.Config{
		BaseURL: apiAddr,
		Token:   getToken(),
	})

	result, err := c.CreateToken(&keysdk.CreateTokenOptions{
		Name:      name,
		Type:      tokenType,
		ExpiresIn: expiresIn,
	})
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		os.Exit(1)
		return
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
}

func tokenSave(cmd *cobra.Command, args []string) {
	tokenValue := args[0]
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		os.Exit(1)
		return
	}

	tokenPath := filepath.Join(homeDir, ".key-agent", "token")
	if err := os.MkdirAll(filepath.Dir(tokenPath), 0700); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		os.Exit(1)
		return
	}

	if err := os.WriteFile(tokenPath, []byte(tokenValue), 0600); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		os.Exit(1)
		return
	}

	fmt.Printf("Token saved to %s\n", tokenPath)
}

// getToken returns the token from flag or file.
func getToken() string {
	if token != "" {
		return token
	}

	// Try to read from specified token file
	if tokenFile != "" {
		data, err := os.ReadFile(tokenFile)
		if err == nil {
			return string(data)
		}
		return ""
	}

	// Try to read from default token file
	homeDir, _ := os.UserHomeDir()
	tokenPath := filepath.Join(homeDir, ".key-agent", "token")
	data, err := os.ReadFile(tokenPath)
	if err == nil {
		return string(data)
	}

	return ""
}
