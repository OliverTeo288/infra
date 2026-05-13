// Package ec2 wraps the AWS EC2 API calls used by the CLI's portforward flow:
// listing instances for selection, and starting an SSM port-forwarding session
// (delegated to the AWS CLI, which has the session-manager-plugin needed).
package ec2

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
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const apiTimeout = 30 * time.Second

// FetchInstances returns display-friendly EC2 instance entries:
//
//	"<instance-id> - <name|(No Name)> [<state>]"
func FetchInstances(profile, region string) ([]string, error) {
	cfg, err := awscfg.LoadConfig(profile, region)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()
	client := ec2.NewFromConfig(cfg)

	var out []string
	paginator := ec2.NewDescribeInstancesPaginator(client, &ec2.DescribeInstancesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe EC2 instances: %w", err)
		}
		for _, r := range page.Reservations {
			for _, inst := range r.Instances {
				out = append(out, fmt.Sprintf("%s - %s [%s]",
					aws.ToString(inst.InstanceId),
					nameFromTags(inst.Tags),
					string(inst.State.Name),
				))
			}
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no EC2 instances available")
	}
	return out, nil
}

// SelectInstance prompts the user to pick an EC2 instance and returns its ID.
func SelectInstance(profile, region string) (string, error) {
	instances, err := FetchInstances(profile, region)
	if err != nil {
		return "", err
	}
	selected, err := prompt.Selection(instances, "EC2 Instance")
	if err != nil {
		return "", err
	}
	return strings.Fields(selected)[0], nil
}

// StartSSMSession starts a port-forwarding SSM session via the AWS CLI.
// Requires session-manager-plugin to be installed on the user's machine.
func StartSSMSession(instanceID, profile, dbHost, region string, dbPort int) error {
	localPort, err := prompt.LocalPort()
	if err != nil {
		return err
	}

	fmt.Printf("Starting SSM session with instance ID: %s\n", instanceID)

	cmd := exec.Command("aws", "ssm", "start-session",
		"--target", instanceID,
		"--document-name", "AWS-StartPortForwardingSessionToRemoteHost",
		"--parameters", fmt.Sprintf(`{"host":["%s"],"portNumber":["%d"],"localPortNumber":["%d"]}`, dbHost, dbPort, localPort),
		"--profile", profile, "--region", region,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start SSM session: %w", err)
	}

	fmt.Println("SSM session started successfully.")
	return nil
}

func nameFromTags(tags []ec2types.Tag) string {
	for _, t := range tags {
		if aws.ToString(t.Key) == "Name" {
			if v := aws.ToString(t.Value); v != "" {
				return v
			}
		}
	}
	return "(No Name)"
}
