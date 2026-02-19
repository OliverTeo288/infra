package functions

import (
	"fmt"
	
	"raid/infra/internal/ecs"
	"raid/infra/internal/utils"
)

func ExecuteECSExec() error {
	// Step 1: Login to AWS
	selectedProfile, selectedRegion, err := utils.Login()
	if err != nil {
		return err
	}

	fmt.Printf("Login successful!\n")

	// Step 2: Select ECS cluster
	cluster, err := ecs.SelectECSCluster(selectedProfile, selectedRegion)
	if err != nil {
		return err
	}

	// Step 3: Select ECS service
	service, err := ecs.SelectECSService(cluster, selectedProfile, selectedRegion)
	if err != nil {
		return err
	}

	// Step 4: Select ECS task
	taskID, err := ecs.SelectECSTask(cluster, service, selectedProfile, selectedRegion)
	if err != nil {
		return err
	}

	// Step 5: Select ECS container
	containerName, err := ecs.SelectECSContainer(cluster, taskID, selectedProfile, selectedRegion)
	if err != nil {
		return err
	}

	fmt.Printf("Cluster: %s, Service: %s, Task ID: %s, Container: %s\n", cluster, service, taskID, containerName)

	// Step 6: Start ECS exec session
	err = ecs.StartECSExecSession(selectedProfile, cluster, taskID, containerName, selectedRegion)
	if err != nil {
		return err
	}

	return nil
}
