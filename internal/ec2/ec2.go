package ec2

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	
	"raid/infra/internal/utils"
)

// Retrieves a list of EC2 instances with their instance IDs and names.
func FetchEC2Instances(profile, region string) ([]string, error) {
	cmd := exec.Command("aws", "ec2", "describe-instances",
		"--query", "Reservations[].Instances[].[InstanceId, Tags[?Key=='Name'].Value | [0], State.Name]",
		"--output", "text",
		"--profile", profile,
		"--region", region)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch EC2 instances: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 {
		fmt.Println("No EC2 instances found.")
		return nil, fmt.Errorf("no EC2 instances available")
	}

	var instances []string
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			instanceID := fields[0]
			instanceName := strings.Join(fields[1:len(fields)-1], " ")
			instanceState := fields[len(fields)-1]
			instances = append(instances, fmt.Sprintf("%s - %s [%s]", instanceID, instanceName, instanceState))
		} else if len(fields) == 2 {
			instanceID := fields[0]
			instanceState := fields[1]
			instances = append(instances, fmt.Sprintf("%s - (No Name) [%s]", instanceID, instanceState))
		}
	}

	return instances, nil
}
// Prompts the user to select an EC2 instance by its ID and Name.
func SelectEC2Instance(profile, region string) (string, error) {
	instances, err := FetchEC2Instances(profile, region)
	if err != nil {
		return "", err
	}

	selectedInstance, err := utils.PromptSelection(instances)
	if err != nil {
		return "", err
	}

	// Extract the instance ID from the selection
	instanceID := strings.Fields(selectedInstance)[0]
	return instanceID, nil
}

// StartSSMSession starts an SSM session with the selected EC2 instance.
func StartEC2SSMSession(instanceID, profile, dbHost, region string, dbPort int) error {
	localPort, err := utils.PromptLocalPortNumber()
	if err != nil {
		return err
	}

	fmt.Printf("Starting SSM session with instance ID: %s\n", instanceID)

		// Run the AWS CLI command to start the SSM session
	cmd := exec.Command("aws", "ssm", "start-session",
		"--target", instanceID,
		"--document-name", "AWS-StartPortForwardingSessionToRemoteHost",
		"--parameters", fmt.Sprintf(`{"host":["%s"],"portNumber":["%d"],"localPortNumber":["%d"]}`, dbHost, dbPort ,localPort),
		"--profile", profile, "--region", region)


	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start SSM session: %v", err)
	}

	fmt.Println("SSM session started successfully.")
	return nil
}
