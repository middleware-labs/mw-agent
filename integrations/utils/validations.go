package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/manifoldco/promptui"
)

var endPointRegex = regexp.MustCompile(`^(?:https?:\/\/)?(localhost|0\.0\.0\.0|(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?|(?:\d{1,3}\.){3}\d{1,3}):[1-9]\d{0,4}(\/[^\s]*(\?[^\s]*)?)?$`)

func PromptWithValidate(label string, validateFunc promptui.ValidateFunc, mask rune) (string, error) {
	prompt := promptui.Prompt{
		Label:    label,
		Validate: validateFunc,
	}
	if mask != 0 {
		prompt.Mask = mask
	}
	return prompt.Run()
}

func ValidateNotEmpty(input string) error {
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("value cannot be empty")
	}
	return nil
}

func ValidateEndpoint(input string) error {
	if !endPointRegex.MatchString(input) {
		return fmt.Errorf("invalid endpoint format")
	}
	return nil
}
