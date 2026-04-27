package utils

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func Login() (string, string, error) {
	profiles, err := getFilteredProfiles()
	if err != nil {
		return "", "", fmt.Errorf("error fetching AWS profiles: %v", err)
	}

	if len(profiles) == 0 {
		fmt.Println("No AWS profiles found. Please configure a new profile using 'aws configure sso'.")
		cmd := exec.Command("aws", "configure", "sso")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", "", fmt.Errorf("failed to configure AWS SSO: %v", err)
		}
		return "", "", nil
	}

	selectedProfile, err := PromptSelection(profiles, "AWS Profile")
	if err != nil {
		return "", "", fmt.Errorf("error selecting AWS profile: %v", err)
	}

	if err := handleExpiredCredentials(selectedProfile); err != nil {
		return "", "", fmt.Errorf("error handling expired credentials for profile '%s': %v", selectedProfile, err)
	}

	selectedRegion, err := FetchAndPromptRegion(selectedProfile)
	if err != nil {
		return "", "", fmt.Errorf("error selecting AWS region: %v", err)
	}

	return selectedProfile, selectedRegion, nil
}

func getFilteredProfiles() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	configPath := filepath.Join(homeDir, ".aws", "config")
	f, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read AWS config: %v", err)
	}
	defer f.Close()

	var profiles []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[profile ") && strings.HasSuffix(line, "]") {
			profileName := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "[profile "), "]"))
			if profileName != "" {
				profiles = append(profiles, profileName)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse AWS config: %v", err)
	}

	sort.Strings(profiles)
	return profiles, nil
}

func handleExpiredCredentials(profile string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "aws", "sts", "get-caller-identity", "--profile", profile)
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
