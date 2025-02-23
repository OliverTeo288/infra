package functions

import (
	"fmt"
	
	"raid/infra/internal/ec2"
	"raid/infra/internal/ecs"
	"raid/infra/internal/rds"
	"raid/infra/internal/utils"
)

// Main logic for port forwarding
func ExecutePortForwarding() error {

	// Step 1: Login to AWS
	selectedProfile, selectedRegion, err := utils.Login()
	if err != nil {
		return err
	}

	fmt.Printf("Login successful!\n")
	
	// // Step 2: Fetch RDS instances and endpoint
	dbHost, dbPort, err := rds.GetRDSEndpoint(selectedProfile, selectedRegion)
	if err != nil {
		return err
	}
	fmt.Printf("Database Host: %s\nPort: %d\n", dbHost, dbPort)

	// Step 3: Prompt user to select EC2 or ECS
	options := []string{"EC2", "ECS"}
	selection, err := utils.PromptSelection(options)
	if err != nil {
		return err
	}

	// Step 4: Execute based on selection
	if selection == "EC2" {
		// EC2 Port Forwarding
		instanceID, err := ec2.SelectEC2Instance(selectedProfile, selectedRegion)
		if err != nil {
			return err
		}
		fmt.Printf("Selected Instance ID: %s\n", instanceID)

		err = ec2.StartEC2SSMSession(instanceID, selectedProfile, dbHost, selectedRegion, dbPort)
		if err != nil {
			return err
		}
	} else if selection == "ECS" {
		// ECS Port Forwarding
		cluster, err := ecs.SelectECSCluster(selectedProfile, selectedRegion)
		if err != nil {
			return err
		}
		service, err := ecs.SelectECSService(cluster, selectedProfile, selectedRegion)
		if err != nil {
			return err
		}
		taskID, err := ecs.SelectECSTask(cluster, service, selectedProfile, selectedRegion)
		if err != nil {
			return err
		}
		runtimeID, err := ecs.GetTaskDetails(cluster, taskID, selectedProfile, selectedRegion)
		if err != nil {
			return err
		}
		containerName, err := ecs.SelectECSContainer(cluster, taskID, selectedProfile, selectedRegion)
		if err != nil {
			return err
		}
		fmt.Printf("Cluster: %s, Service: %s, Task ID: %s, Runtime ID: %s, Container: %s\n", cluster, service, taskID, runtimeID, containerName)

		err = ecs.StartECSSSMSession(selectedProfile, cluster, taskID, runtimeID, dbHost, selectedRegion, dbPort)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("invalid selection: %s", selection)
	}
	
	return nil
}