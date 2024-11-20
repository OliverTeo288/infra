package rds

import (
	"fmt"
	"os/exec"
	"encoding/json"
	"strings"
	"github.com/oliverteo288/infra/internal/utils"
)

// GetRDSInstance fetches the RDS instance identifier for a given profile
func GetRDSInstance(profile, region string) (string, error) {
	cmd := exec.Command("aws", "rds", "describe-db-instances", "--query", "DBInstances[].DBInstanceIdentifier", "--output", "text", "--profile", profile, "--region", region)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to fetch RDS instances: %v", err)
	}

	dbs := strings.Fields(strings.TrimSpace(string(output)))

	if len(dbs) == 0 {
		fmt.Println("No RDS instances found.")
		return "", fmt.Errorf("no RDS instances found")
	}

	fmt.Println("Select an RDS Instance:")
	for i, db := range dbs {
		fmt.Printf("[%d] %s\n", i+1, db)
	}

	return utils.PromptSelection(dbs)
}

// GetRDSInstanceEndpoint fetches the endpoint of the RDS instance
func GetRDSInstanceEndpoint(identifier, profile, region string) (string, int, error) {

	cmd := exec.Command("aws", "rds", "describe-db-instances",
		"--db-instance-identifier", identifier,
		"--query", "DBInstances[0].Endpoint",
		"--output", "json",
		"--profile", profile,
		"--region", region,
	)

	output, err := cmd.Output()
	if err != nil {
		return "", 0, fmt.Errorf("failed to fetch RDS endpoint: %v", err)
	}

	var endpoint struct {
		Address string `json:"Address"`
		Port    int    `json:"Port"`
	}
	if err := json.Unmarshal(output, &endpoint); err != nil {
		return "", 0, fmt.Errorf("failed to parse RDS endpoint JSON: %v", err)
	}

	return endpoint.Address, endpoint.Port, nil
}