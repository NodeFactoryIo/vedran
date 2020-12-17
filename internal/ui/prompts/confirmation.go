package prompts

import (
	"fmt"
	"github.com/manifoldco/promptui"
)

// ShowSelectionPrompt displays confirmation dialog with label describing what user is prompted
// and enables user to choose between options.
func ShowSelectionPrompt(label string, options []string) (bool, error) {
	prompt := promptui.Select{
		Label: label,
		Items: options,
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return false, err
	}

	// fmt.Printf("You choose %q\n", result)
	return "Yes" == result, nil
}

// ShowSelectionPrompt displays confirmation dialog with label describing what user is prompted
// and enables user to choose between "Yes" and "No".
func ShowConfirmationPrompt(label string) (bool, error) {
	return ShowSelectionPrompt(label, []string{"Yes", "No"})
}
