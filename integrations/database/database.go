package database

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/middleware-labs/mw-agent/integrations/services"
	"github.com/middleware-labs/mw-agent/integrations/utils"
)

func ConfigurePostgres(hostname string, agentConfig utils.AgentConfig) {
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

	services.SaveAndSendConfig("postgres", config, hostname, agentConfig)
}

func ConfigureMongoDB(hostname string, agentConfig utils.AgentConfig) {
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

	services.SaveAndSendConfig("mongodb", config, hostname, agentConfig)
}

func ConfigureOracleDB(hostname string, agentConfig utils.AgentConfig) {
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

	services.SaveAndSendConfig("oracledb", config, hostname, agentConfig)
}

func ConfigureMySQL(hostname string, agentConfig utils.AgentConfig) {
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

	services.SaveAndSendConfig("mysql", config, hostname, agentConfig)
}

func ConfigureRedis(hostname string, agentConfig utils.AgentConfig) {
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

	services.SaveAndSendConfig("redis", config, hostname, agentConfig)
}

func ConfigureMariaDB(hostname string, agentConfig utils.AgentConfig) {
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

	services.SaveAndSendConfig("mariadb", config, hostname, agentConfig)
}

func ConfigureElasticSearch(hostname string, agentConfig utils.AgentConfig) {
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

	services.SaveAndSendConfig("elasticsearch", config, hostname, agentConfig)
}
