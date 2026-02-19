package utils

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func Login() (string, string, error) {
	// Fetch all available AWS profiles
	profiles, err := getFilteredProfiles()
	if err != nil {
		return "", "",fmt.Errorf("error fetching AWS profiles: %v", err)
	}

	// Handle empty profile list by guiding the user to configure a new profile
	if len(profiles) == 0 {
		fmt.Println("No AWS profiles found. Please configure a new profile using 'aws configure sso'.")
		cmd := exec.Command("aws", "configure", "sso")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", "", fmt.Errorf("failed to configure AWS SSO: %v", err)
		}
		return "", "",nil
	}

	// Prompt user to select a profile
	selectedProfile, err := PromptSelection(profiles, "AWS Profile")
	if err != nil {
		return "", "",fmt.Errorf("error selecting AWS profile: %v", err)
	}

	fmt.Printf("Selected AWS Profile: %s\n", selectedProfile)

	// Handle expired credentials for the selected profile
	if err := handleExpiredCredentials(selectedProfile); err != nil {
		return "", "", fmt.Errorf("error handling expired credentials for profile '%s': %v", selectedProfile, err)
	}

	// Fetch and prompt for region selection
	selectedRegion, err := FetchAndPromptRegion(selectedProfile)
	if err != nil {
		return "", "", fmt.Errorf("error selecting AWS region: %v", err)
	}

	fmt.Printf("Selected AWS Region: %s\n", selectedRegion)
	fmt.Println("AWS profile and region.")

	return selectedProfile, selectedRegion, nil
}



func getFilteredProfiles() ([]string, error) {
	cmd := exec.Command("sh", "-c", `grep '\[profile' ~/.aws/config | sed 's/\[profile \(.*\)\]/\1/'`)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch AWS profiles: %v", err)
	}
	profiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	sort.Strings(profiles)
	return profiles, nil
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

