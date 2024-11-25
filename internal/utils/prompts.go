package utils

import (
	"bufio"
	"fmt"
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Prompts the user to select from a list of options
func PromptSelection(options []string) (string, error) {
	attempts := 0

	for attempts < 3 {
		// Display the options for user reference
		fmt.Println("Please select from the following options:")
		for i, option := range options {
			fmt.Printf("[%d] %s\n", i+1, option)
		}

		fmt.Print("Enter the number of your choice: ")
		reader := bufio.NewReader(os.Stdin)
		choice, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Failed to read input: %v\n", err)
			attempts++
			continue
		}

		index := strings.TrimSpace(choice)
		if i, err := strconv.Atoi(index); err == nil && i > 0 && i <= len(options) {
			return options[i-1], nil
		}

		attempts++
		fmt.Printf("Invalid choice. You have %d attempt(s) remaining.\n", 3-attempts)
	}

	// If the user fails 3 times, exit with an error
	return "", fmt.Errorf("too many invalid attempts")
}

// Prompts the user for a local port number for port forwarding
func PromptLocalPortNumber() (int, error) {
	reader := bufio.NewReader(os.Stdin)
	attempts := 0

	for attempts < 3 {
		fmt.Println("Enter a local port number for port forwarding (1024–65535):")
		input, err := reader.ReadString('\n')
		if err != nil {
			return 0, fmt.Errorf("failed to read input: %v", err)
		}

		// Trim and parse the input
		input = strings.TrimSpace(input)
		port, err := strconv.Atoi(input)
		if err == nil && port >= 1024 && port <= 65535 {
			return port, nil
		}

		fmt.Println("Invalid port number. Please enter a number between 1024 and 65535.")
		attempts++
	}

	return 0, fmt.Errorf("too many invalid attempts. Exiting.")
}

// Fetches available regions using the AWS CLI and prompts the user to select one.
func FetchAndPromptRegion(profile string) (string, error) {
	// Run AWS CLI command to fetch regions
	cmd := exec.Command("aws", "ec2", "describe-regions", "--profile", profile, "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to fetch regions: %v", err)
	}

	// Parse JSON output
	var regions struct {
		Regions []struct {
			RegionName string `json:"RegionName"`
		} `json:"Regions"`
	}
	if err := json.Unmarshal(output, &regions); err != nil {
		return "", fmt.Errorf("failed to parse regions JSON: %v", err)
	}

	// Extract region names
	regionNames := []string{}
	for _, region := range regions.Regions {
		regionNames = append(regionNames, region.RegionName)
	}

	return PromptSelection(regionNames)
}

func PromptInput(prompt string, validate func(input string) error, defaultValue string) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	// Store the original prompt string for reuse
	basePrompt := prompt
	if defaultValue != "" {
		basePrompt = fmt.Sprintf("%s [%s]: ", prompt, defaultValue)
	} else {
		basePrompt = fmt.Sprintf("%s: ", prompt)
	}

	for {
		// Display the prompt
		fmt.Print(basePrompt)

		// Read user input
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		// Trim any leading/trailing whitespace
		input = strings.TrimSpace(input)

		// Use default value if input is empty
		if input == "" && defaultValue != "" {
			input = defaultValue
		}

		// Validate input if a validation function is provided
		if validate != nil {
			if err := validate(input); err != nil {
				fmt.Printf("Invalid input: %s. Please try again.\n", err)
				continue
			}
		}

		return input, nil
	}
}

// confirmPrompt displays a confirmation prompt and returns true if the user confirms.
func ConfirmPrompt(message string) bool {
	fmt.Print(message + " ")
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		fmt.Println("Error reading input:", err)
		return false
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y"
}