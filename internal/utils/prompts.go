package utils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// PromptSelection prompts the user to select from a list of options
func PromptSelection(options []string) (string, error) {
	fmt.Print("Enter the number of your choice: ")
	reader := bufio.NewReader(os.Stdin)
	choice, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %v", err)
	}
	index := strings.TrimSpace(choice)
	if i, err := strconv.Atoi(index); err == nil && i > 0 && i <= len(options) {
		return options[i-1], nil
	}
	return "", fmt.Errorf("invalid choice")
}

// PromptLocalPortNumber prompts the user for a local port number for port forwarding
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