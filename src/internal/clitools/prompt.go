package clitools

import (
	"fmt"
	"os"
	"path/filepath"

	survey "github.com/AlecAivazis/survey/v2"
)

func AskPath(kind, def string, predefined []string) (string, error) {
	var (
		path            string
		err             error
		shouldCreateNew bool
	)
	createNewPrompt := &survey.Confirm{
		Message: fmt.Sprintf("Do you want to create a new %s?", kind),
	}
	if err = survey.AskOne(createNewPrompt, &shouldCreateNew); err != nil {
		return "", fmt.Errorf("failed to ask if new %s should be created: %w", kind, err)
	}
	if !shouldCreateNew {
		return def, nil // NOTE: this is intentional
	}

	if def != "" {
		path = def
	} else {
		path, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	// Add an option for the user to enter a custom path
	predefined = append(predefined, "Enter custom path")

	prompt := &survey.Select{
		Message: fmt.Sprintf("Choose %s path:", kind),
		Options: predefined,
	}

	err = survey.AskOne(prompt, &path)
	if err != nil {
		return "", fmt.Errorf("failed to ask for %s path: %w", kind, err)
	}

	if path == "Enter custom path" {
		prompt := &survey.Input{
			Message: "Enter the custom path:",
		}
		err = survey.AskOne(prompt, &path)
		if err != nil {
			return "", fmt.Errorf("failed to ask for custom path: %w", err)
		}
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return path, nil
}
