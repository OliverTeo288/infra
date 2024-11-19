package utils

import (
	"bufio"
	"fmt"
	// "context"
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	// "github.com/aws/aws-sdk-go-v2/config"
	// "github.com/aws/aws-sdk-go-v2/service/ec2"
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

	// Prompt user to select a region
	fmt.Println("Select an AWS Region:")
	for i, region := range regionNames {
		fmt.Printf("[%d] %s\n", i+1, region)
	}
	return PromptSelection(regionNames)
}

// func FetchAndPromptRegion(profile string) (string, error) {
// 	// Load AWS configuration with the selected profile
// 	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(profile))
// 	if err != nil {
// 		return "", fmt.Errorf("failed to load AWS config: %v", err)
// 	}

// 	// Create an EC2 client
// 	client := ec2.NewFromConfig(cfg)

// 	// Fetch the list of regions
// 	resp, err := client.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
// 	if err != nil {
// 		return "", fmt.Errorf("failed to fetch regions: %v", err)
// 	}

// 	// Extract region names
// 	regionNames := []string{}
// 	for _, region := range resp.Regions {
// 		regionNames = append(regionNames, *region.RegionName)
// 	}

// 	// Prompt user to select a region
// 	fmt.Println("Select an AWS Region:")
// 	for i, region := range regionNames {
// 		fmt.Printf("[%d] %s\n", i+1, region)
// 	}
// 	return PromptSelection(regionNames)
// }

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