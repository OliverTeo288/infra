package rds

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"raid/infra/internal/utils"
)

var enginePortMap = map[string]int{
	"MYSQL":      3306,
	"POSTGRESQL": 5432,
	"MARIADB":    3306,
	"SQLSERVER":  1433,
	"ORACLE":     1521,
}

// GetRDSEndpoint fetches the endpoint and port for an RDS instance or proxy
func GetRDSEndpoint(profile, region string) (string, int, error) {
	selections, err := fetchRDSSelections(profile, region)
	if err != nil {
		return "", 0, err
	}

	identifier, err := utils.PromptSelection(selections, "RDS Instance or Proxy")
	if err != nil {
		return "", 0, err
	}

	if strings.HasPrefix(identifier, "[RDS proxy]") {
		return fetchProxyEndpoint(identifier, profile, region)
	}
	return fetchInstanceEndpoint(identifier, profile, region)
}

// fetchRDSSelections gathers both RDS instances and proxies
func fetchRDSSelections(profile, region string) ([]string, error) {
	instances, err := fetchRDSInstances(profile, region)
	if err != nil {
		return nil, err
	}

	proxies, err := fetchRDSProxies(profile, region)
	if err != nil {
		fmt.Printf("Note: could not fetch RDS proxies: %v\n", err)
	}
	selections := append(instances, proxies...)

	if len(selections) == 0 {
		return nil, errors.New("no RDS instances or proxies found")
	}
	return selections, nil
}

func fetchRDSInstances(profile, region string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "aws", "rds", "describe-db-instances", "--query", "DBInstances[].DBInstanceIdentifier", "--output", "text", "--profile", profile, "--region", region)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("failed to fetch RDS instances: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("failed to fetch RDS instances: %v", err)
	}

	instances := strings.Fields(strings.TrimSpace(string(output)))
	for i, db := range instances {
		instances[i] = fmt.Sprintf("[RDS instance] %s", db)
	}
	return instances, nil
}

func fetchRDSProxies(profile, region string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "aws", "rds", "describe-db-proxies", "--query", "DBProxies[].DBProxyName", "--output", "text", "--profile", profile, "--region", region)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("%s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, err
	}

	proxies := strings.Fields(strings.TrimSpace(string(output)))
	for i, proxy := range proxies {
		proxies[i] = fmt.Sprintf("[RDS proxy] %s", proxy)
	}
	return proxies, nil
}

func fetchInstanceEndpoint(identifier, profile, region string) (string, int, error) {
	identifier = strings.TrimPrefix(identifier, "[RDS instance] ")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "aws", "rds", "describe-db-instances", "--db-instance-identifier", identifier, "--query", "DBInstances[0].Endpoint", "--output", "json", "--profile", profile, "--region", region)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", 0, fmt.Errorf("failed to fetch instance endpoint: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", 0, fmt.Errorf("failed to fetch instance endpoint: %v", err)
	}

	var endpoint struct {
		Address string `json:"Address"`
		Port    int    `json:"Port"`
	}
	if err := json.Unmarshal(output, &endpoint); err != nil {
		return "", 0, fmt.Errorf("failed to parse instance endpoint JSON: %v", err)
	}
	return endpoint.Address, endpoint.Port, nil
}

func fetchProxyEndpoint(identifier, profile, region string) (string, int, error) {
	identifier = strings.TrimPrefix(identifier, "[RDS proxy] ")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "aws", "rds", "describe-db-proxies", "--db-proxy-name", identifier, "--query", "DBProxies[0].[Endpoint, EngineFamily]", "--output", "json", "--profile", profile, "--region", region)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", 0, fmt.Errorf("failed to fetch proxy endpoint: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", 0, fmt.Errorf("failed to fetch proxy endpoint: %v", err)
	}

	var proxyResult []interface{}
	if err := json.Unmarshal(output, &proxyResult); err != nil {
		return "", 0, fmt.Errorf("failed to parse proxy endpoint JSON: %v", err)
	}
	if len(proxyResult) != 2 {
		return "", 0, fmt.Errorf("unexpected proxy endpoint format")
	}

	address, ok := proxyResult[0].(string)
	if !ok || address == "" {
		return "", 0, fmt.Errorf("invalid or missing proxy endpoint address")
	}
	engineFamily, ok := proxyResult[1].(string)
	if !ok || engineFamily == "" {
		return "", 0, fmt.Errorf("invalid or missing proxy engine family")
	}
	port, ok := enginePortMap[strings.ToUpper(engineFamily)]
	if !ok {
		return "", 0, fmt.Errorf("unknown engine family %q", engineFamily)
	}
	return address, port, nil
}
