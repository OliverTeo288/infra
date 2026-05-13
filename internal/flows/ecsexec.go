package flows

import (
	"context"

	"raid/infra/internal/awsauth"
	"raid/infra/internal/ecs"
	"raid/infra/internal/observability"
)

// ECSExec drives the full `infra ecs exec` workflow: pick profile/region/
// cluster/service/task/container, then open an interactive shell via ECS
// Exec.
func ECSExec(ctx context.Context, shell string) error {
	profile, region, err := awsauth.Login(ctx)
	if err != nil {
		return err
	}

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

	observability.FromContext(ctx).Info("ecs_exec_session_starting",
		"cluster", cluster, "service", service, "task_id", taskID,
		"container", containerName, "shell", shell,
	)

	return ecs.StartExecSession(profile, cluster, taskID, containerName, region, shell)
}
