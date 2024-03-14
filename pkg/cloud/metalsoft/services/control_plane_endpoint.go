package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/metalsoft-io/cluster-api-provider-metalsoft/pkg/cloud/metalsoft"
	"github.com/metalsoft-io/cluster-api-provider-metalsoft/pkg/cloud/metalsoft/scope"
	metalcloud "github.com/metalsoft-io/metal-cloud-sdk-go/v2"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type ControlPlaneEndpointService struct {
	Client                *metalsoft.MetalSoftClient
	InfrastructureService *InfrastructureService
	SubnetService         *SubnetService
	VariablesService      *VariablesService
}

func NewControlPlaneEndpointService(client *metalsoft.MetalSoftClient) *ControlPlaneEndpointService {
	infrastructureService := &InfrastructureService{
		client,
	}

	subnetService := &SubnetService{
		client,
	}

	variablesService := &VariablesService{
		client,
	}

	return &ControlPlaneEndpointService{
		Client:                client,
		InfrastructureService: infrastructureService,
		SubnetService:         subnetService,
		VariablesService:      variablesService,
	}
}

func (es *ControlPlaneEndpointService) GetEndpoint(ctx context.Context, clusterScope *scope.ClusterScope) (string, error) {
	log := log.FromContext(ctx)

	var infrastructure *metalcloud.Infrastructure
	var subnet *metalcloud.Subnet
	var err error

	datacenterName := clusterScope.DatacenterName()
	infrastructureLabel := clusterScope.InfrastructureLabel()
	vipSubnetLabel := ""
	// vipSubnetLabel := clusterScope.VipSubnetLabel()
	infrastructureID := clusterScope.InfrastructureID()
	subnetId := clusterScope.SubnetID()
	subnetSubdomain := clusterScope.SubnetSubdomain()

	if datacenterName == "" {
		return "", errors.New("DatacenterName is required")
	}

	if infrastructureLabel == "" {
		infrastructureLabel = "cluster-api-" + generateRandomID()
	}

	if vipSubnetLabel == "" {
		vipSubnetLabel = "cluster-api-subnet-" + generateRandomID()
	}

	if infrastructureID != 0 {
		infrastructure, err = es.InfrastructureService.getInfrastructure(infrastructureID)

		// 	// TODO: Differentiate if the infrastructure is part of a cluster or not // check infrastructure.operation

		if err != nil {
			return "", errors.Wrap(err, "failed to get infrastructure")
		}
	} else {
		infrastructure, err = es.InfrastructureService.createGetInfrastructure(infrastructureLabel, datacenterName)

		if err != nil {
			return "", errors.Wrap(err, "failed to create or get existing infrastructure")
		}
		clusterScope.SetInfrastructureID(infrastructure.InfrastructureID)
	}

	if subnetId != 0 {
		subnet, err = es.SubnetService.getSubnet(subnetId)

		if err != nil {
			return "", errors.Wrap(err, "failed to get subnet")
		}
	} else {
		networks, err := es.Client.Networks(infrastructure.InfrastructureID)

		if err != nil {
			log.Error(err, "Error getting networks")
			return "", err
		}

		wanNetworkId := (*networks)["wan"].NetworkID

		if wanNetworkId == 0 {
			log.Error(err, "Error getting wan network")
			return "", errors.New("wan network not found")
		}

		subnet, err = es.SubnetService.createGetSubnet(wanNetworkId, infrastructure.InfrastructureID, vipSubnetLabel)

		if err != nil {
			log.Error(err, "Error creating or getting subnet")
			return "", errors.Wrap(err, "failed to create or get existing subnet")
		}

		clusterScope.SetSubnetID(subnet.SubnetID)
	}

	if subnet.SubnetSubdomain == "" {
		log.Info("SubnetSubdomain not found")
		return "", errors.New("subnetSubDomain not found")
	}

	if subnetSubdomain == "" {
		// We are setting the variable kube_vip_address with the subnetSubdomain only once
		clusterScope.SetSubnetSubdomain(subnet.SubnetSubdomain)

		variableName := "kube_vip_address"

		variableValue := `{"value":" ` + subnetSubdomain + ` "}`
		_, err = es.VariablesService.createVariable(variableName, variableValue)

		if err != nil {
			return "", errors.Wrap(err, "failed to create or get existing variable")
		}
	}

	return subnet.SubnetSubdomain, nil
}

func generateRandomID() string {
	return uuid.New().String()
}
