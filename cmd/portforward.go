/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"github.com/oliverteo288/infra/internal/ecs"
  "github.com/oliverteo288/infra/internal/rds"
  "github.com/oliverteo288/infra/internal/utils"
	"github.com/spf13/cobra"
)

// portforwardCmd represents the portforward command
var portforwardCmd = &cobra.Command{
	Use:   "portforward",
	Short: "Making it easier for you to portfoward into your Private RDS from your ECS",
	Long: `Automatically discovers your ECS Tasks and RDS hostname to start an SSM session for DB management. Please ensure that you have a profile already set in place`,
	Run: func(cmd *cobra.Command, args []string) {
		err := executePortForwarding()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(portforwardCmd)
}

// Main logic for port forwarding
func executePortForwarding() error {

	// Step 1: Login to AWS
	selectedProfile, selectedRegion, err := utils.Login()
	if err != nil {
		return err
	}

	fmt.Printf("Login successful! %s\n", selectedRegion)

	// Step 2: Fetch RDS instances and prompt user for a DB selection
	dbIdentifier, err := rds.GetRDSInstance(selectedProfile, selectedRegion)
	if err != nil {
		return err
	}
	fmt.Printf("Using DB Identifier: %s\n", dbIdentifier)

	// Step 3: Discover ECS clusters, services, tasks, and containers
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

	// Step 4: Fetch RDS endpoint
	dbHost, dbPort ,err := rds.GetRDSInstanceEndpoint(dbIdentifier, selectedProfile, selectedRegion)
	if err != nil {
		return err
	}
	fmt.Printf("Database Host: %s\n", dbHost)

	// Step 5: Start SSM session
	err = ecs.StartSSMSession(selectedProfile, cluster, taskID, runtimeID, dbHost, selectedRegion, dbPort)
	if err != nil {
		return err
	}

	fmt.Println("Port forwarding session started successfully.")
	return nil
}