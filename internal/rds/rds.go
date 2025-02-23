package rds

import (
	"fmt"
	"os/exec"
	"encoding/json"
	"strings"
	"errors"
	
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

	identifier, err := utils.PromptSelection(selections)
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

	proxies, _ := fetchRDSProxies(profile, region)
	selections := append(instances, proxies...)

	if len(selections) == 0 {
		return nil, errors.New("no RDS instances or proxies found")
	}
	return selections, nil
}

func fetchRDSInstances(profile, region string) ([]string, error) {
	cmd := exec.Command("aws", "rds", "describe-db-instances", "--query", "DBInstances[].DBInstanceIdentifier", "--output", "text", "--profile", profile, "--region", region)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RDS instances: %v", err)
	}

	instances := strings.Fields(strings.TrimSpace(string(output)))
	for i, db := range instances {
		instances[i] = fmt.Sprintf("[RDS instance] %s", db)
	}
	return instances, nil
}

func fetchRDSProxies(profile, region string) ([]string, error) {
	cmd := exec.Command("aws", "rds", "describe-db-proxies", "--query", "DBProxies[].DBProxyName", "--output", "text", "--profile", profile, "--region", region)
	output, err := cmd.Output()
	if err != nil {
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
	cmd := exec.Command("aws", "rds", "describe-db-instances", "--db-instance-identifier", identifier, "--query", "DBInstances[0].Endpoint", "--output", "json", "--profile", profile, "--region", region)
	
	output, err := cmd.Output()
	if err != nil {
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
	cmd := exec.Command("aws", "rds", "describe-db-proxies", "--db-proxy-name", identifier, "--query", "DBProxies[0].[Endpoint, EngineFamily]", "--output", "json", "--profile", profile, "--region", region)
	
	output, err := cmd.Output()
	if err != nil {
		return "", 0, fmt.Errorf("failed to fetch proxy endpoint: %v", err)
	}

	var proxyResult []interface{}
	if err := json.Unmarshal(output, &proxyResult); err != nil {
		return "", 0, fmt.Errorf("failed to parse proxy endpoint JSON: %v", err)
	}
	if len(proxyResult) != 2 {
		return "", 0, fmt.Errorf("unexpected proxy endpoint format")
	}

	address, _ := proxyResult[0].(string)
	engineFamily, _ := proxyResult[1].(string)
	port := enginePortMap[strings.ToUpper(engineFamily)]
	return address, port, nil
}
