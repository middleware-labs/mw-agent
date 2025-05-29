package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/middleware-labs/mw-agent/integrations/services"
	"github.com/middleware-labs/mw-agent/integrations/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

	saveAndSendConfig("postgres", config, hostname)
}

func ConfigureMongoDB(hostname string) {
	fmt.Println("\nYou selected: MongoDB")

	var endpoints []map[string]string

	for {
		// Prompt for each MongoDB endpoint
		endpoint, err := utils.PromptWithValidate("Enter MongoDB Endpoint (e.g., localhost:27017)", utils.ValidateEndpoint, 0)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		endpoints = append(endpoints, map[string]string{
			"endpoint": endpoint,
		})

		// Ask to add another endpoint
		addAnotherPrompt := &promptui.Select{
			Label: "Add another endpoint?",
			Items: []string{"Yes", "No"},
		}
		addAnother, _, err := addAnotherPrompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed: %v\n", err)
			return
		}
		if addAnother == 1 {
			break
		}
	}

	// Username
	username, err := utils.PromptWithValidate("Enter Username", utils.ValidateNotEmpty, 0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Password
	password, err := utils.PromptWithValidate("Enter Password", utils.ValidateNotEmpty, '*')
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Final config structure
	config := map[string]map[string]interface{}{
		"mongodb": {
			"hosts":    endpoints,
			"username": username,
			"password": password,
		},
	}

	saveAndSendConfig("mongodb", config, hostname)
}

func ConfigureOracleDB(hostname string) {
	fmt.Println("\nYou selected: OracleDB")

	endpoint, err := utils.PromptWithValidate("Enter Endpoint (e.g., localhost:1521)", utils.ValidateEndpoint, 0)
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
		"oracledb": {
			"endpoint": endpoint,
			"username": username,
			"password": password,
		},
	}

	saveAndSendConfig("oracledb", config, hostname)
}

func ConfigureMySQL(hostname string) {
	fmt.Println("\nYou selected: MySQL")

	endpoint, err := utils.PromptWithValidate("🔌 Enter Endpoint (e.g., localhost:3306)", utils.ValidateEndpoint, 0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	username, err := utils.PromptWithValidate("👤 Enter Username", utils.ValidateNotEmpty, 0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	password, err := utils.PromptWithValidate("🔑 Enter Password", utils.ValidateNotEmpty, '*')
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	config := map[string]map[string]string{
		"mysql": {
			"endpoint": endpoint,
			"username": username,
			"password": password,
		},
	}

	saveAndSendConfig("mysql", config, hostname)
}

func ConfigureRedis(hostname string) {
	fmt.Println("\nYou selected: Redis")

	endpoint, err := utils.PromptWithValidate("Enter Redis Endpoint (e.g., localhost:6379)", utils.ValidateEndpoint, 0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	config := map[string]map[string]string{
		"redis": {
			"endpoint": endpoint,
		},
	}

	saveAndSendConfig("redis", config, hostname)
}

func ConfigureMariaDB(hostname string) {
	fmt.Println("\nYou selected: MariaDB")

	endpoint, err := utils.PromptWithValidate("Enter Endpoint (e.g., localhost:3306)", utils.ValidateEndpoint, 0)
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
		"mariadb": {
			"endpoint": endpoint,
			"username": username,
			"password": password,
		},
	}

	saveAndSendConfig("mariadb", config, hostname)
}

func ConfigureElasticSearch(hostname string) {
	fmt.Println("\nYou selected: ElasticSearch")

	endpoint, err := utils.PromptWithValidate("Enter Endpoint (e.g., http://localhost:9200)", utils.ValidateEndpoint, 0)
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
		"elasticsearch": {
			"endpoint": endpoint,
			"username": username,
			"password": password,
		},
	}

	saveAndSendConfig("elasticsearch", config, hostname)
}

func saveAndSendConfig(integration string, config interface{}, hostname string) {
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		fmt.Printf("Failed to marshal YAML: %v\n", err)
		return
	}

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

	filePath := filepath.Join(baseDir, integration+".yaml")
	err = os.WriteFile(filePath, yamlData, 0644)
	if err != nil {
		fmt.Printf("Failed to write config file: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Printf("🎉 %s configuration completed!\n", cases.Title(language.English).String(integration))
	fmt.Printf("📁 File saved at: %s", filePath)

	baseURL := ""      // TODO: Add the base URL of your API (Local or Deployed)
	accountToken := "" // TODO: API key for authentication

	authData, err := services.FetchAuthData(baseURL, hostname, accountToken)
	if err != nil {
		fmt.Printf("❌ Auth fetch error: %v\n", err)
		return
	}

	services.SendIntegrationConfigToAPI(baseURL, filePath, hostname, authData, integration+"_config")

	fmt.Println("──────────────────────────────────────────────")
	color.Green("✅ Setup complete! You may now monitor %s.", cases.Title(language.English).String(integration))
	fmt.Println("──────────────────────────────────────────────")

}
