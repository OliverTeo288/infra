package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)


func Login() (string, error) {
	// Fetch all available AWS profiles
	profiles, err := getFilteredProfiles()
	if err != nil {
		return "", fmt.Errorf("error fetching AWS profiles: %v", err)
	}

	// Handle empty profile list by guiding the user to configure a new profile
	if len(profiles) == 0 {
		fmt.Println("No AWS profiles found. Please configure a new profile using 'aws configure sso'.")
		cmd := exec.Command("aws", "configure", "sso")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to configure AWS SSO: %v", err)
		}
		return "", nil
	}

	// Prompt user to select a profile
	selectedProfile, err := promptProfileSelection(profiles)
	if err != nil {
		return "", fmt.Errorf("error selecting AWS profile: %v", err)
	}

	fmt.Printf("Selected AWS Profile: %s\n", selectedProfile)

	// Handle expired credentials for the selected profile
	if err := handleExpiredCredentials(selectedProfile); err != nil {
		return "", fmt.Errorf("error handling expired credentials for profile '%s': %v", selectedProfile, err)
	}

	fmt.Println("AWS profile is ready to use.")
	return selectedProfile, nil
}

func getFilteredProfiles() ([]string, error) {
	cmd := exec.Command("sh", "-c", `grep '\[profile' ~/.aws/config | sed 's/\[profile \(.*\)\]/\1/'`)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch AWS profiles: %v", err)
	}
	profiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	return profiles, nil
}

func promptProfileSelection(profiles []string) (string, error) {
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

func handleExpiredCredentials(profile string) error {
	// Check if AWS CLI command fails due to expired credentials
	cmd := exec.Command("aws", "sts", "get-caller-identity", "--profile", profile)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Credentials for profile '%s' have expired or are invalid. Logging in...\n", profile)
		loginCmd := exec.Command("aws", "sso", "login", "--profile", profile)
		loginCmd.Stdout = os.Stdout
		loginCmd.Stderr = os.Stderr
		err = loginCmd.Run()
		if err != nil {
			return fmt.Errorf("failed to log in to AWS SSO: %v", err)
		}
		fmt.Println("AWS SSO login successful.")
	}
	return nil
}