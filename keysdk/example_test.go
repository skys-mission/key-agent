package keysdk_test

import (
	"fmt"
	"log"

	keysdk "github.com/skys-mission/key-agent/keysdk"
)

func ExampleNewClient() {
	// Create a new client
	client := keysdk.NewClient(&keysdk.Config{
		BaseURL: "http://127.0.0.1:8080",
		Token:   "your-token-here",
	})

	// Check server health
	health, err := client.Health()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Server status: %s, version: %s\n", health.Status, health.Version)
}

func Example_kvOperations() {
	client := keysdk.NewClient(&keysdk.Config{
		Token: "your-token-here",
	})

	// Set a KV entry
	entry, err := client.SetKV("my-key", &keysdk.SetKVOptions{
		Value: "my-value",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created: %s\n", entry.Key)

	// Get a KV entry
	entry, err = client.GetKV("my-key")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Value: %s\n", entry.Value)

	// List KV keys
	keys, err := client.ListKV("")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Keys: %v\n", keys)

	// Delete a KV entry
	if err := client.DeleteKV("my-key"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deleted")
}

func Example_secretOperations() {
	client := keysdk.NewClient(&keysdk.Config{
		Token: "your-token-here",
	})

	// Set a secret
	secret, err := client.SetSecret("db-password", &keysdk.SetSecretOptions{
		Value: "super-secret-password",
		Type:  keysdk.SecretTypePassword,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created secret: %s\n", secret.Key)

	// Get a secret
	secret, err = client.GetSecret("db-password")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Secret value: %s\n", secret.Value)
}

func ExampleClient_GetStringOr() {
	client := keysdk.NewClient(&keysdk.Config{
		BaseURL: "http://127.0.0.1:8080",
		Token:   "your-token-here",
	})

	// Get KV value with default fallback
	debugMode := client.GetStringOr("app/debug", "false")
	fmt.Printf("Debug mode: %s\n", debugMode)

	// If key exists, returns actual value
	// If key not found or error, returns default "false"
}

func ExampleClient_GetSecretStringOr() {
	client := keysdk.NewClient(&keysdk.Config{
		BaseURL: "http://127.0.0.1:8080",
		Token:   "your-token-here",
	})

	// Get secret value with default fallback
	dbPassword := client.GetSecretStringOr("db/password", "")
	if dbPassword == "" {
		fmt.Println("Using default empty password")
	}

	// If secret exists, returns actual value
	// If secret not found or error, returns default ""
}

func Example_tokenOperations() {
	client := keysdk.NewClient(&keysdk.Config{
		Token: "your-root-token",
	})

	// Create a new token
	token, err := client.CreateToken(&keysdk.CreateTokenOptions{
		Name:      "my-app",
		Type:      "client",
		ExpiresIn: "24h",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Token: %s\n", token.Token)
}