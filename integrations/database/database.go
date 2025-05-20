package database

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v2"
)

var endPointRegex = regexp.MustCompile(`^(?:https?:\/\/)?(localhost|0\.0\.0\.0|(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?|(?:\d{1,3}\.){3}\d{1,3}):[1-9]\d{0,4}(\/[^\s]*(\?[^\s]*)?)?$`)

func promptWithValidate(label string, validateFunc promptui.ValidateFunc, mask rune) (string, error) {
	prompt := promptui.Prompt{
		Label:    label,
		Validate: validateFunc,
	}
	if mask != 0 {
		prompt.Mask = mask
	}
	return prompt.Run()
}

func validateNotEmpty(input string) error {
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("value cannot be empty")
	}
	return nil
}

func validateEndpoint(input string) error {
	if !endPointRegex.MatchString(input) {
		return fmt.Errorf("invalid endpoint format")
	}
	return nil
}

func ConfigurePostgres() {
	fmt.Println("\nYou selected: Postgres")

	endpoint, err := promptWithValidate("Enter Endpoint (e.g., localhost:5432)", validateEndpoint, 0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	username, err := promptWithValidate("Enter Username", validateNotEmpty, 0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	password, err := promptWithValidate("Enter Password", validateNotEmpty, '*')
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
}
