// Package ecs wraps the AWS ECS API calls used by the CLI for cluster /
// service / task / container selection, and the two interactive session
// flavours (ECS Exec shell, SSM port-forwarding via runtime ID).
package ecs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"raid/infra/internal/awscfg"
	"raid/infra/internal/prompt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

const apiTimeout = 30 * time.Second

// GetClusters returns cluster short-names for the given profile/region.
func GetClusters(profile, region string) ([]string, error) {
	cfg, err := awscfg.LoadConfig(profile, region)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()
	client := ecs.NewFromConfig(cfg)

	var names []string
	paginator := ecs.NewListClustersPaginator(client, &ecs.ListClustersInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list ECS clusters: %w", err)
		}
		for _, arn := range page.ClusterArns {
			names = append(names, lastSegment(arn))
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no ECS clusters available")
	}
	return names, nil
}

// SelectCluster prompts the user to pick an ECS cluster.
func SelectCluster(profile, region string) (string, error) {
	clusters, err := GetClusters(profile, region)
	if err != nil {
		return "", err
	}
	return prompt.Selection(clusters, "ECS Cluster")
}

// GetServices returns service short-names within a cluster.
func GetServices(cluster, profile, region string) ([]string, error) {
	cfg, err := awscfg.LoadConfig(profile, region)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()
	client := ecs.NewFromConfig(cfg)

	var names []string
	paginator := ecs.NewListServicesPaginator(client, &ecs.ListServicesInput{
		Cluster: aws.String(cluster),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list ECS services: %w", err)
		}
		for _, arn := range page.ServiceArns {
			names = append(names, lastSegment(arn))
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no ECS services available")
	}
	return names, nil
}

// SelectService prompts the user to pick an ECS service.
func SelectService(cluster, profile, region string) (string, error) {
	services, err := GetServices(cluster, profile, region)
	if err != nil {
		return "", err
	}
	return prompt.Selection(services, "ECS Service")
}

// GetTasks returns task IDs for a given service.
func GetTasks(cluster, service, profile, region string) ([]string, error) {
	cfg, err := awscfg.LoadConfig(profile, region)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()
	client := ecs.NewFromConfig(cfg)

	var ids []string
	paginator := ecs.NewListTasksPaginator(client, &ecs.ListTasksInput{
		Cluster:     aws.String(cluster),
		ServiceName: aws.String(service),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list ECS tasks: %w", err)
		}
		for _, arn := range page.TaskArns {
			ids = append(ids, lastSegment(arn))
		}
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("no ECS tasks available")
	}
	return ids, nil
}

// SelectTask prompts the user to pick an ECS task.
func SelectTask(cluster, service, profile, region string) (string, error) {
	tasks, err := GetTasks(cluster, service, profile, region)
	if err != nil {
		return "", err
	}
	return prompt.Selection(tasks, "ECS Task")
}

// GetRuntimeID returns the runtimeId for the named container within the given task.
func GetRuntimeID(cluster, taskID, containerName, profile, region string) (string, error) {
	cfg, err := awscfg.LoadConfig(profile, region)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()
	client := ecs.NewFromConfig(cfg)

	out, err := client.DescribeTasks(ctx, &ecs.DescribeTasksInput{
		Cluster: aws.String(cluster),
		Tasks:   []string{taskID},
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe ECS task: %w", err)
	}
	if len(out.Tasks) == 0 {
		return "", fmt.Errorf("task %s not found in cluster %s", taskID, cluster)
	}
	for _, c := range out.Tasks[0].Containers {
		if aws.ToString(c.Name) == containerName {
			rt := aws.ToString(c.RuntimeId)
			if rt == "" {
				return "", fmt.Errorf("container %s has no runtime ID (task may not be running)", containerName)
			}
			return rt, nil
		}
	}
	return "", fmt.Errorf("container %s not found in task %s", containerName, taskID)
}

// GetContainers returns container names for the given task.
func GetContainers(cluster, taskID, profile, region string) ([]string, error) {
	cfg, err := awscfg.LoadConfig(profile, region)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()
	client := ecs.NewFromConfig(cfg)

	out, err := client.DescribeTasks(ctx, &ecs.DescribeTasksInput{
		Cluster: aws.String(cluster),
		Tasks:   []string{taskID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe ECS task: %w", err)
	}
	if len(out.Tasks) == 0 {
		return nil, fmt.Errorf("task %s not found in cluster %s", taskID, cluster)
	}
	var names []string
	for _, c := range out.Tasks[0].Containers {
		names = append(names, aws.ToString(c.Name))
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no containers found for task %s", taskID)
	}
	return names, nil
}

// SelectContainer prompts the user to pick a container within a task.
func SelectContainer(cluster, taskID, profile, region string) (string, error) {
	containers, err := GetContainers(cluster, taskID, profile, region)
	if err != nil {
		return "", err
	}
	return prompt.Selection(containers, "ECS Container")
}

// StartExecSession opens an interactive ECS Exec session via the AWS CLI.
// shell defaults to /bin/sh when empty.
func StartExecSession(profile, cluster, taskID, containerName, region, shell string) error {
	if shell == "" {
		shell = "/bin/sh"
	}
	cmd := exec.Command("aws", "ecs", "execute-command",
		"--cluster", cluster,
		"--task", taskID,
		"--container", containerName,
		"--interactive",
		"--command", shell,
		"--profile", profile,
		"--region", region,
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start ECS exec session: %w", err)
	}
	return nil
}

// StartSSMSession starts a port-forwarding SSM session targeting an ECS
// container via the AWS CLI.
func StartSSMSession(profile, cluster, taskID, runtimeID, dbHost, region string, dbPort int) error {
	localPort, err := prompt.LocalPort()
	if err != nil {
		return err
	}

	target := fmt.Sprintf("ecs:%s_%s_%s", cluster, taskID, runtimeID)
	fmt.Printf("SSM Target: %s\n", target)

	cmd := exec.Command("aws", "ssm", "start-session",
		"--target", target,
		"--document-name", "AWS-StartPortForwardingSessionToRemoteHost",
		"--parameters", fmt.Sprintf(`{"host":["%s"],"portNumber":["%d"],"localPortNumber":["%d"]}`, dbHost, dbPort, localPort),
		"--profile", profile, "--region", region,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start SSM session: %w", err)
	}
	return nil
}

// lastSegment returns the substring after the final "/" in an ARN-like string.
func lastSegment(arn string) string {
	parts := strings.Split(arn, "/")
	return parts[len(parts)-1]
}
