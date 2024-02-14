package metalsoft

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	metalcloud "github.com/metalsoft-io/metal-cloud-sdk-go/v2"
	"github.com/pkg/errors"
	// "metalsoft.local/metalsoft"
)

const (
	infraLabelEnvName = "INFRASTRUCTURE_LABEL"
)

var (
	infraLabel string
	idRegex    = regexp.MustCompile(`ID (\d+)`)
)

const (
	infrastructureAlreadyExistsRefCode = "d3241ebb479ffdd886b4dca6a61e3263"
	subnetAlreadyExistsRefCode         = "96cb5b6dffa2dcd7805e02e89fd890cb"
	variableAlreadyExistsRefCode       = "d5cf17eb3a9ad5bf460c1d564b47d395"
)

type MetalsoftClusterSpec struct {
	InfrastructureLabel string `json:"infrastructureLabel"`
	DatacenterName      string `json:"datacenterName"`
	VipSubnetLabel      string `json:"vipSubnetLabel"`
}

func setControlPlaneEndpoint(client *metalcloud.Client, datacenterName string, infraLabel string, vipSubnetLabel string) (string, error) {
	infraDataFromCluster := MetalsoftClusterSpec{
		InfrastructureLabel: infraLabel,
		DatacenterName:      datacenterName,
		VipSubnetLabel:      vipSubnetLabel,
	}

	if infraDataFromCluster.DatacenterName == "" {
		return "", errors.New("DatacenterName is required")
	}

	if infraDataFromCluster.InfrastructureLabel == "" {
		infraDataFromCluster.InfrastructureLabel = "cluster-api-" + generateRandomID()
	}

	if infraDataFromCluster.VipSubnetLabel == "" {
		infraDataFromCluster.VipSubnetLabel = "cluster-api-subnet-" + generateRandomID()
	}

	infraObject := metalcloud.Infrastructure{
		InfrastructureLabel: infraDataFromCluster.InfrastructureLabel,
		DatacenterName:      infraDataFromCluster.DatacenterName,
	}

	infra, err := client.InfrastructureCreate(infraObject)

	if err != nil {
		fmt.Printf("Error creating infrastructure: %v\n", err)
		return "", err
	}

	// TODO: Differentiate if the infrastructure is part of a cluster or not // check infrastructure.operation

	networks, err := client.Networks(infra.InfrastructureID)

	if err != nil {
		fmt.Printf("Error getting networks: %v\n", err)
		return "", err
	}

	wanNetworkId := (*networks)["wan"].NetworkID

	if wanNetworkId == 0 {
		fmt.Printf("Error getting wan network: %v\n", err)
		return "", errors.New("wan network not found")
	}

	subnet, err := createGetSubnet(client, wanNetworkId, infra.InfrastructureID, infraDataFromCluster.VipSubnetLabel)

	if err != nil {
		fmt.Printf("Error creating subnet: %v\n", err)
		return "", err
	}
	subnetSubDomain := subnet.SubnetSubdomainPermanent

	if subnetSubDomain == "" {
		fmt.Printf("Error getting subnetSubDomain: %v\n", err)
		return "", errors.New("subnetSubDomain not found")
	}

	variableName := "kube_vip_address"
	variableValue := `{"value":" ` + subnetSubDomain + ` "}`

	variable, err := createGetVariable(client, variableName, variableValue)

	if err != nil {
		fmt.Printf("Error creating variable: %v\n", err)
		return "", err
	}

	return variable.VariableName, nil
}

func createGetInfrastructure(client *metalcloud.Client, spec MetalsoftClusterSpec) (*metalcloud.Infrastructure, error) {
	infra := metalcloud.Infrastructure{
		InfrastructureLabel: spec.InfrastructureLabel,
		DatacenterName:      spec.DatacenterName,
	}
	createdInfra, err := client.InfrastructureCreate(infra)
	if err != nil {
		if strings.Contains(err.Error(), infrastructureAlreadyExistsRefCode) {
			id, err := extractIDFromError(err.Error())
			if err != nil {
				return nil, err
			}
			return client.InfrastructureGet(id)
		}
		return nil, errors.Wrap(err, "failed to create or get existing infrastructure")
	}
	return createdInfra, nil
}

func createGetSubnet(client *metalcloud.Client, wanNetworkId int, infraID int, subnetLabel string) (*metalcloud.Subnet, error) {
	subnet := metalcloud.Subnet{
		NetworkID:                 wanNetworkId,
		InfrastructureID:          infraID,
		SubnetLabel:               subnetLabel,
		SubnetDestination:         "wan",
		SubnetPrefixSize:          29, // TODO: make this 30
		SubnetType:                "ipv4",
		SubnetAutomaticAllocation: false,
	}
	createdSubnet, err := client.SubnetCreate(subnet)
	if err != nil {
		if strings.Contains(err.Error(), subnetAlreadyExistsRefCode) {
			id, err := extractIDFromError(err.Error())
			if err != nil {
				return nil, err
			}
			return client.SubnetGet(id)
		}
		return nil, errors.Wrap(err, "failed to create or get existing subnet")
	}
	return createdSubnet, nil
}

func createGetVariable(client *metalcloud.Client, variableName, variableValue string) (*metalcloud.Variable, error) {
	variable := metalcloud.Variable{
		VariableName: variableName,
		VariableJSON: variableValue,
	}
	createdVariable, err := client.VariableCreate(variable)
	if err != nil {
		if strings.Contains(err.Error(), variableAlreadyExistsRefCode) {
			// Handle existing variable scenario, potentially updating it or simply retrieving it
			return &variable, nil
		}
		return nil, errors.Wrap(err, "failed to create variable")
	}
	return createdVariable, nil
}

func generateRandomID() string {
	return uuid.New().String()
}

func extractIDFromError(err string) (int, error) {
	matches := idRegex.FindStringSubmatch(err)
	if len(matches) < 2 {
		return 0, fmt.Errorf("no ID found in error string")
	}
	id, convErr := strconv.Atoi(matches[1])
	if convErr != nil {
		return 0, errors.Wrap(convErr, "failed to convert ID to integer")
	}
	return id, nil
}
