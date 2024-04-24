package services

import (
	"strings"

	"github.com/metalsoft-io/cluster-api-provider-metalsoft/pkg/cloud/metalsoft"
	metalcloud "github.com/metalsoft-io/metal-cloud-sdk-go/v3"
	"github.com/pkg/errors"
)

type SubnetService struct {
	Client *metalsoft.MetalSoftClient
}

func NewSubnetService(client *metalsoft.MetalSoftClient) *SubnetService {
	return &SubnetService{
		Client: client,
	}
}

func (service *SubnetService) CreateGetSubnet(wanNetworkId int, infraID int, subnetLabel string) (*metalcloud.Subnet, error) {
	subnet := metalcloud.Subnet{
		NetworkID:                 wanNetworkId,
		InfrastructureID:          infraID,
		SubnetLabel:               subnetLabel,
		SubnetDestination:         "wan",
		SubnetPrefixSize:          29, // TODO: make this 30 | Err: Subnet exhausted trying to allocate 2 system reserved IPs.
		SubnetType:                "ipv4",
		SubnetAutomaticAllocation: false,
	}

	createdSubnet, err := service.Client.SubnetCreate(subnet)

	if err != nil {
		if strings.Contains(err.Error(), subnetAlreadyExistsRefCode) || strings.Contains(err.Error(), "42bde79f729065b21e8583dd00cf48e0") {
			id, err := extractIDFromError(err.Error())
			if err != nil {
				return nil, err
			}
			return service.GetSubnet(id)
		}
		return nil, errors.Wrap(err, "failed to create or get existing subnet")
	}

	return createdSubnet, nil
}

func (service *SubnetService) GetSubnet(subnetID int) (*metalcloud.Subnet, error) {
	subnet, err := service.Client.SubnetGet(subnetID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get subnet")
	}
	return subnet, nil
}
