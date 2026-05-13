package flows

import (
	"context"
	"fmt"

	"raid/infra/internal/awsauth"
	"raid/infra/internal/ec2"
	"raid/infra/internal/ecs"
	"raid/infra/internal/observability"
	"raid/infra/internal/prompt"
	"raid/infra/internal/rds"
)

// targetEC2, targetECS are the two compute hosts the SSM port-forward
// document can target.
const (
	targetEC2 = "EC2"
	targetECS = "ECS"
)

// PortForward drives the full `infra portforward` workflow:
//
//  1. Pick AWS profile and region
//  2. Resolve an RDS endpoint
//  3. Pick a compute host (EC2 or ECS task) to relay through
//  4. Start the SSM port-forwarding session
func PortForward(ctx context.Context) error {
	log := observability.FromContext(ctx)

	profile, region, err := awsauth.Login(ctx)
	if err != nil {
		return err
	}

	dbHost, dbPort, err := rds.GetEndpoint(profile, region)
	if err != nil {
		return err
	}
	log.Info("rds_endpoint_resolved", "host", dbHost, "port", dbPort)

	target, err := prompt.Selection([]string{targetEC2, targetECS}, "")
	if err != nil {
		return err
	}

	switch target {
	case targetEC2:
		return portForwardViaEC2(ctx, profile, region, dbHost, dbPort)
	case targetECS:
		return portForwardViaECS(ctx, profile, region, dbHost, dbPort)
	default:
		return fmt.Errorf("invalid selection: %s", target)
	}
}

func portForwardViaEC2(ctx context.Context, profile, region, dbHost string, dbPort int) error {
	instanceID, err := ec2.SelectInstance(profile, region)
	if err != nil {
		return err
	}
	observability.FromContext(ctx).Info("ec2_target_selected", "instance_id", instanceID)
	return ec2.StartSSMSession(instanceID, profile, dbHost, region, dbPort)
}

func portForwardViaECS(ctx context.Context, profile, region, dbHost string, dbPort int) error {
	cluster, err := ecs.SelectCluster(profile, region)
	if err != nil {
		return err
	}
	service, err := ecs.SelectService(cluster, profile, region)
	if err != nil {
		return err
	}
	taskID, err := ecs.SelectTask(cluster, service, profile, region)
	if err != nil {
		return err
	}
	containerName, err := ecs.SelectContainer(cluster, taskID, profile, region)
	if err != nil {
		return err
	}
	runtimeID, err := ecs.GetRuntimeID(cluster, taskID, containerName, profile, region)
	if err != nil {
		return err
	}
	observability.FromContext(ctx).Info("ecs_target_selected",
		"cluster", cluster, "service", service, "task_id", taskID,
		"container", containerName, "runtime_id", runtimeID,
	)
	return ecs.StartSSMSession(profile, cluster, taskID, runtimeID, dbHost, region, dbPort)
}
