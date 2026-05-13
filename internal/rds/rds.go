// Package rds wraps the AWS RDS API for listing instances/proxies and
// resolving their endpoints into a (host, port) pair the SSM port-forward
// document can target.
package rds

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"raid/infra/internal/awscfg"
	"raid/infra/internal/prompt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

const apiTimeout = 30 * time.Second

const (
	rdsInstancePrefix = "[RDS instance] "
	rdsProxyPrefix    = "[RDS proxy] "
)

// enginePortMap maps RDS proxy EngineFamily values to default DB ports.
var enginePortMap = map[string]int{
	"MYSQL":      3306,
	"POSTGRESQL": 5432,
	"MARIADB":    3306,
	"SQLSERVER":  1433,
	"ORACLE":     1521,
}

// GetEndpoint prompts the user to pick an RDS instance or proxy, then returns
// its (host, port) tuple.
func GetEndpoint(profile, region string) (string, int, error) {
	cfg, err := awscfg.LoadConfig(profile, region)
	if err != nil {
		return "", 0, err
	}
	client := rds.NewFromConfig(cfg)

	selections, err := fetchSelections(client)
	if err != nil {
		return "", 0, err
	}

	identifier, err := prompt.Selection(selections, "RDS Instance or Proxy")
	if err != nil {
		return "", 0, err
	}

	if strings.HasPrefix(identifier, rdsProxyPrefix) {
		return proxyEndpoint(client, strings.TrimPrefix(identifier, rdsProxyPrefix))
	}
	return instanceEndpoint(client, strings.TrimPrefix(identifier, rdsInstancePrefix))
}

func fetchSelections(client *rds.Client) ([]string, error) {
	instances, err := listInstances(client)
	if err != nil {
		return nil, err
	}

	// Proxy listing is best-effort: not every account uses RDS Proxy and we
	// don't want a Proxy IAM denial to break instance access.
	proxies, err := listProxies(client)
	if err != nil {
		fmt.Printf("Note: could not fetch RDS proxies: %v\n", err)
	}

	all := append(instances, proxies...)
	if len(all) == 0 {
		return nil, errors.New("no RDS instances or proxies found")
	}
	return all, nil
}

func listInstances(client *rds.Client) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	var out []string
	paginator := rds.NewDescribeDBInstancesPaginator(client, &rds.DescribeDBInstancesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe RDS instances: %w", err)
		}
		for _, db := range page.DBInstances {
			out = append(out, rdsInstancePrefix+aws.ToString(db.DBInstanceIdentifier))
		}
	}
	return out, nil
}

func listProxies(client *rds.Client) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	var out []string
	paginator := rds.NewDescribeDBProxiesPaginator(client, &rds.DescribeDBProxiesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe RDS proxies: %w", err)
		}
		for _, p := range page.DBProxies {
			out = append(out, rdsProxyPrefix+aws.ToString(p.DBProxyName))
		}
	}
	return out, nil
}

func instanceEndpoint(client *rds.Client, identifier string) (string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	out, err := client.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(identifier),
	})
	if err != nil {
		return "", 0, fmt.Errorf("failed to fetch instance endpoint: %w", err)
	}
	if len(out.DBInstances) == 0 || out.DBInstances[0].Endpoint == nil {
		return "", 0, fmt.Errorf("instance %s has no endpoint", identifier)
	}
	ep := out.DBInstances[0].Endpoint
	return aws.ToString(ep.Address), int(aws.ToInt32(ep.Port)), nil
}

func proxyEndpoint(client *rds.Client, identifier string) (string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	out, err := client.DescribeDBProxies(ctx, &rds.DescribeDBProxiesInput{
		DBProxyName: aws.String(identifier),
	})
	if err != nil {
		return "", 0, fmt.Errorf("failed to fetch proxy endpoint: %w", err)
	}
	if len(out.DBProxies) == 0 {
		return "", 0, fmt.Errorf("proxy %s not found", identifier)
	}
	p := out.DBProxies[0]
	address := aws.ToString(p.Endpoint)
	if address == "" {
		return "", 0, fmt.Errorf("proxy %s has no endpoint", identifier)
	}
	engineFamily := aws.ToString(p.EngineFamily)
	port, ok := enginePortMap[strings.ToUpper(engineFamily)]
	if !ok {
		return "", 0, fmt.Errorf("unknown engine family %q for proxy %s", engineFamily, identifier)
	}
	return address, port, nil
}
