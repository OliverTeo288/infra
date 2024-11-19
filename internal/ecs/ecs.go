package ecs

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"github.com/oliverteo288/infra/internal/utils"
)

// GetECSClusters fetches the ECS clusters for a given profile
func GetECSClusters(profile, region string) ([]string, error) {
	cmd := exec.Command("aws", "ecs", "list-clusters", "--query", "clusterArns", "--output", "text", "--profile", profile, "--region", region)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ECS clusters: %v", err)
	}

	arns := strings.Fields(strings.TrimSpace(string(output)))
	if len(arns) == 0 {
		fmt.Println("No ECS clusters found.")
		return nil, fmt.Errorf("no ECS clusters available")
	}

	var clusters []string

	for _, arn := range arns {
		parts := strings.Split(arn, "/")
		clusters = append(clusters, parts[len(parts)-1])
	}
	return clusters, nil
}

// SelectECSCluster prompts the user to select an ECS cluster
func SelectECSCluster(profile, region string) (string, error) {
	clusters, err := GetECSClusters(profile, region)
	if err != nil {
		return "", err
	}
	fmt.Println("Select an ECS Cluster:")
	for i, cluster := range clusters {
		fmt.Printf("[%d] %s\n", i+1, cluster)
	}
	return utils.PromptSelection(clusters) // Call the promptSelection function from utils
}

// GetECSServices fetches the ECS services for a given cluster and profile
func GetECSServices(cluster, profile, region string) ([]string, error) {
	cmd := exec.Command("aws", "ecs", "list-services", "--cluster", cluster, "--query", "serviceArns", "--output", "text", "--profile", profile, "--region", region)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ECS services: %v", err)
	}
	arns := strings.Fields(strings.TrimSpace(string(output)))

	if len(arns) == 0 {
		fmt.Println("No ECS services found.")
		return nil, fmt.Errorf("no ECS services available")
	}
	var services []string
	for _, arn := range arns {
		parts := strings.Split(arn, "/")
		services = append(services, parts[len(parts)-1])
	}
	return services, nil
}

// SelectECSService prompts the user to select an ECS service
func SelectECSService(cluster, profile, region string) (string, error) {
	services, err := GetECSServices(cluster, profile, region)
	if err != nil {
		return "", err
	}
	fmt.Println("Select an ECS Service:")
	for i, service := range services {
		fmt.Printf("[%d] %s\n", i+1, service)
	}
	return utils.PromptSelection(services) // Call the promptSelection function from utils
}

// GetECSTasks fetches the ECS tasks for a given cluster, service, and profile
func GetECSTasks(cluster, service, profile, region string) ([]string, error) {
	cmd := exec.Command("aws", "ecs", "list-tasks", "--cluster", cluster, "--service-name", service, "--query", "taskArns", "--output", "text", "--profile", profile, "--region", region)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ECS tasks: %v", err)
	}
	arns := strings.Fields(strings.TrimSpace(string(output)))

	if len(arns) == 0 {
		fmt.Println("No ECS tasks found.")
		return nil, fmt.Errorf("no ECS tasks available")
	}
	var taskIDs []string
	for _, arn := range arns {
		parts := strings.Split(arn, "/")
		taskIDs = append(taskIDs, parts[len(parts)-1]) // Extract task ID
	}
	return taskIDs, nil
}

// SelectECSTask prompts the user to select an ECS task
func SelectECSTask(cluster, service, profile, region string) (string, error) {
	tasks, err := GetECSTasks(cluster, service, profile, region)
	if err != nil {
		return "", err
	}
	fmt.Println("Select an ECS Task:")
	for i, task := range tasks {
		fmt.Printf("[%d] %s\n", i+1, task)
	}
	return utils.PromptSelection(tasks) // Call the promptSelection function from utils
}

// GetTaskDetails fetches the details for an ECS task
func GetTaskDetails(cluster, taskID, profile, region string) (string, error) {
	cmd := exec.Command("aws", "ecs", "describe-tasks", "--cluster", cluster, "--tasks", taskID, "--query", "tasks[0].containers[0].runtimeId", "--output", "text", "--profile", profile, "--region", region)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to fetch ECS task details: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetECSContainers fetches the container names for a given task
func GetECSContainers(cluster, taskID, profile, region string) ([]string, error) {
	cmd := exec.Command("aws", "ecs", "describe-tasks", "--cluster", cluster, "--tasks", taskID, "--query", "tasks[0].containers[].name", "--output", "text", "--profile", profile, "--region", region)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ECS containers: %v", err)
	}
	containers := strings.Fields(strings.TrimSpace(string(output)))

	if len(containers) == 0 {
		fmt.Println("No ECS containers found.")
		return nil, fmt.Errorf("no ECS containers available")
	}
	return containers, nil
}

// SelectECSContainer prompts the user to select an ECS container
func SelectECSContainer(cluster, taskID, profile, region string) (string, error) {
	containers, err := GetECSContainers(cluster, taskID, profile, region)
	if err != nil {
		return "", err
	}
	fmt.Println("Select an ECS Container:")
	for i, container := range containers {
		fmt.Printf("[%d] %s\n", i+1, container)
	}
	return utils.PromptSelection(containers) // Call the promptSelection function from utils
}

// StartSSMSession starts an SSM session for port forwarding
func StartSSMSession(profile, cluster, taskID, runtimeID, dbHost, region string) error {
	// Prompt user for a local port number
	localPort, err := utils.PromptLocalPortNumber()
	if err != nil {
		return err
	}

	
	target := fmt.Sprintf("ecs:%s_%s_%s", cluster, taskID, runtimeID)
	fmt.Printf("SSM Target: %s\n", target)

	// Run the AWS CLI command to start the SSM session
	cmd := exec.Command("aws", "ssm", "start-session",
		"--target", target,
		"--document-name", "AWS-StartPortForwardingSessionToRemoteHost",
		"--parameters", fmt.Sprintf(`{"host":["%s"],"portNumber":["5432"],"localPortNumber":["%d"]}`, dbHost, localPort),
		"--profile", profile, "--region", region)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to start SSM session: %v", err)
	}
	return nil
}