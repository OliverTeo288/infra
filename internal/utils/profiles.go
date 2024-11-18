package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// getFilteredProfiles fetches AWS profiles from the AWS config file
func GetFilteredProfiles() ([]string, error) {
	cmd := exec.Command("sh", "-c", `grep '\[profile' ~/.aws/config | sed 's/\[profile \(.*\)\]/\1/'`)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch AWS profiles: %v", err)
	}
	profiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	return profiles, nil
}

// promptProfileSelection prompts the user to select an AWS profile
func PromptProfileSelection(profiles []string) (string, error) {
	fmt.Println("Select an AWS Profile:")
	for i, profile := range profiles {
		fmt.Printf("[%d] %s\n", i+1, profile)
	}
	fmt.Print("Enter the number of your choice: ")

	reader := bufio.NewReader(os.Stdin)
	choice, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %v", err)
	}

	index := strings.TrimSpace(choice)
	if i, err := strconv.Atoi(index); err == nil && i > 0 && i <= len(profiles) {
		return profiles[i-1], nil
	}
	return "", fmt.Errorf("invalid choice")
}