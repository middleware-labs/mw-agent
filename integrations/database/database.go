package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/middleware-labs/mw-agent/integrations/services"
	"github.com/middleware-labs/mw-agent/integrations/utils"
	"gopkg.in/yaml.v2"
)

func ConfigurePostgres(hostname string) {
	fmt.Println("\nYou selected: Postgres")

	endpoint, err := utils.PromptWithValidate("Enter Endpoint (e.g., localhost:5432)", utils.ValidateEndpoint, 0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	username, err := utils.PromptWithValidate("Enter Username", utils.ValidateNotEmpty, 0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	password, err := utils.PromptWithValidate("Enter Password", utils.ValidateNotEmpty, '*')
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	config := map[string]map[string]string{
		"postgresql": {
			"endpoint": endpoint,
			"username": username,
			"password": password,
		},
	}

	yamlData, err := yaml.Marshal(config)
	if err != nil {
		fmt.Printf("Failed to marshal YAML: %v\n", err)
		return
	}

	// Prepare directory path
	baseDir := "/etc/mw-agent/integrations"
	err = os.MkdirAll(baseDir, 0755)
	if err != nil {
		if os.IsPermission(err) {
			fmt.Println("❌ Permission denied. Please run this command with sudo:")
			fmt.Println("   sudo mw-agent integration")
		} else {
			fmt.Printf("Failed to create directories: %v\n", err)
		}
		return
	}

	filePath := filepath.Join(baseDir, "postgres.yaml")
	err = os.WriteFile(filePath, yamlData, 0644)
	if err != nil {
		fmt.Printf("Failed to write config file: %v\n", err)
		return
	}

	fmt.Println("\n✅ Postgres has been configured and saved successfully!")
	fmt.Printf("📄 Saved to: %s\n", filePath)

	baseURL := ""      // TODO: Add the base URL of your API (Local or Deployed)
	accountToken := "" // TODO: API key for authentication

	authData, err := services.FetchAuthData(baseURL, hostname, accountToken)
	if err != nil {
		fmt.Printf("❌ Auth fetch error: %v\n", err)
		return
	}

	// Call the API service to send the config
	services.SendPostgresConfigToAPI(baseURL, filePath, hostname, authData)
}
