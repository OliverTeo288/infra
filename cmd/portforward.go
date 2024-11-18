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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:
infra portforward`,
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
	// Step 1: Fetch available AWS profiles
	profiles, err := utils.GetFilteredProfiles()
	if err != nil {
		return err
	}

	// Step 2: Prompt user to select a profile
	selectedProfile, err := utils.PromptProfileSelection(profiles)
	if err != nil {
		return err
	}
	fmt.Printf("Selected AWS Profile: %s\n", selectedProfile)

	// Step 3: Fetch RDS instances and prompt user for a DB selection
	dbIdentifier, err := rds.GetRDSInstance(selectedProfile)  // Use internal RDS package
	if err != nil {
		return err
	}
	fmt.Printf("Using DB Identifier: %s\n", dbIdentifier)

	// Step 4: Discover ECS clusters, services, tasks, and containers
	cluster, err := ecs.SelectECSCluster(selectedProfile)  // Use internal ECS package
	if err != nil {
		return err
	}
	service, err := ecs.SelectECSService(cluster, selectedProfile)
	if err != nil {
		return err
	}
	taskID, err := ecs.SelectECSTask(cluster, service, selectedProfile)
	if err != nil {
		return err
	}
	runtimeID, err := ecs.GetTaskDetails(cluster, taskID, selectedProfile)
	if err != nil {
		return err
	}
	containerName, err := ecs.SelectECSContainer(cluster, taskID, selectedProfile)
	if err != nil {
		return err
	}
	fmt.Printf("Cluster: %s, Service: %s, Task ID: %s, Runtime ID: %s, Container: %s\n", cluster, service, taskID, runtimeID, containerName)

	// Step 5: Fetch RDS endpoint
	dbHost, err := rds.GetRDSInstanceEndpoint(dbIdentifier, selectedProfile)  // Use internal RDS package
	if err != nil {
		return err
	}
	fmt.Printf("Database Host: %s\n", dbHost)

	// Step 6: Start SSM session
	err = ecs.StartSSMSession(selectedProfile, cluster, taskID, runtimeID, dbHost)
	if err != nil {
		return err
	}

	fmt.Println("Port forwarding session started successfully.")
	return nil
}