package integrations

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/middleware-labs/mw-agent/integrations/database"
	"github.com/middleware-labs/mw-agent/integrations/utils"
	"gopkg.in/yaml.v2"
)

func RunIntegrationSelection(hostname string) {
	// Select category first
	categories := []string{}
	for cat := range utils.CategorizedIntegrations {
		categories = append(categories, cat)
	}

	categoryPrompt := promptui.Select{
		Label: "📚 Select a category of integrations",
		Items: categories,
		Size:  8,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "👉 {{ . | underline }}",
			Inactive: "   {{ . }}",
			Selected: "🎯 Selected: {{ . }}",
		},
	}

	_, selectedCategory, err := categoryPrompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
		return
	}

	// Now list integrations under that category with numbers
	integrations := utils.CategorizedIntegrations[selectedCategory]
	numberedItems := make([]string, len(integrations))
	for i, val := range integrations {
		numberedItems[i] = fmt.Sprintf("%d. %s", i+1, val)
	}

	integrationPrompt := promptui.Select{
		Label: fmt.Sprintf("🔧 Choose a service in '%s'", selectedCategory),
		Items: numberedItems,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "👉 {{ . | underline }}",
			Inactive: "   {{ . }}",
			Selected: "🎉 Selected: {{ . }}",
		},
		Size: 10,
	}

	idx, _, err := integrationPrompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
		return
	}

	// Use the raw integration name (without the number)
	selectedIntegration := integrations[idx]
	configureIntegration(selectedIntegration, hostname)
}

func configureIntegration(integration, hostname string) {
	color.Cyan("Configuring '%s' for host '%s'...\n", integration, hostname)

	if !utils.IsMwAgentRunning() {
		fmt.Println("mw-agent is not running.")
		return
	}

	data, err := os.ReadFile("/etc/mw-agent/agent-config.yaml")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var config utils.AgentConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("Failed to parse YAML: %v", err)
	}

	switch integration {
	case "Postgres":
		database.ConfigurePostgres(hostname, config)
	case "MySQL":
		database.ConfigureMySQL(hostname, config)
	case "MariaDB":
		database.ConfigureMongoDB(hostname, config)
	case "MongoDB":
		database.ConfigureMongoDB(hostname, config)
	case "OracleDB":
		database.ConfigureOracleDB(hostname, config)
	case "Redis":
		database.ConfigureRedis(hostname, config)
	case "ElasticSearch":
		database.ConfigureElasticSearch(hostname, config)
	default:
		fmt.Printf("\nℹ️ Configuration UI for %s is not implemented yet.\n", integration)
	}
}
