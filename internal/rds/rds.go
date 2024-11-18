package rds

import (
	"fmt"
	"os/exec"
	"strings"
	"github.com/oliverteo288/infra/internal/utils"
)

// GetRDSInstance fetches the RDS instance identifier for a given profile
func GetRDSInstance(profile string) (string, error) {
	cmd := exec.Command("aws", "rds", "describe-db-instances", "--query", "DBInstances[].DBInstanceIdentifier", "--output", "text", "--profile", profile)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to fetch RDS instances: %v", err)
	}
	dbs := strings.Fields(strings.TrimSpace(string(output)))

	fmt.Println("Select an RDS Instance:")
	for i, db := range dbs {
		fmt.Printf("[%d] %s\n", i+1, db)
	}
	return utils.PromptSelection(dbs) // Call the promptSelection function from prompts
}

// GetRDSInstanceEndpoint fetches the endpoint of the RDS instance
func GetRDSInstanceEndpoint(identifier, profile string) (string, error) {
	cmd := exec.Command("aws", "rds", "describe-db-instances", "--db-instance-identifier", identifier, "--query", "DBInstances[0].Endpoint.Address", "--output", "text", "--profile", profile)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to fetch RDS endpoint: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}
