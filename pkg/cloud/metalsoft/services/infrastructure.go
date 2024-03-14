package services

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/metalsoft-io/cluster-api-provider-metalsoft/pkg/cloud/metalsoft"
	metalcloud "github.com/metalsoft-io/metal-cloud-sdk-go/v2"
	"github.com/pkg/errors"
)

type InfrastructureService struct {
	*metalsoft.MetalSoftClient
}

var (
	idRegex = regexp.MustCompile(`ID (\d+)`)
)

const (
	infrastructureAlreadyExistsRefCode = "d3241ebb479ffdd886b4dca6a61e3263"
	subnetAlreadyExistsRefCode         = "96cb5b6dffa2dcd7805e02e89fd890cb"
)

func (service *InfrastructureService) createGetInfrastructure(infrastructureLabel string, datacenterName string) (*metalcloud.Infrastructure, error) {
	infra := metalcloud.Infrastructure{
		InfrastructureLabel: infrastructureLabel,
		DatacenterName:      datacenterName,
	}

	createdInfra, err := service.InfrastructureCreate(infra)

	if err != nil {
		if strings.Contains(err.Error(), infrastructureAlreadyExistsRefCode) {
			id, err := extractIDFromError(err.Error())
			if err != nil {
				return nil, err
			}
			return service.InfrastructureGet(id)
		}
		return nil, errors.Wrap(err, "failed to create or get existing infrastructure")
	}
	return createdInfra, nil
}

func (service *InfrastructureService) getInfrastructure(infrastructureID int) (*metalcloud.Infrastructure, error) {
	infra, err := service.InfrastructureGet(infrastructureID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get infrastructure")
	}
	return infra, nil
}

func extractIDFromError(err string) (int, error) {
	matches := idRegex.FindStringSubmatch(err)
	if len(matches) < 2 {
		return 0, errors.New("failed to extract ID from string")
	}
	id, convErr := strconv.Atoi(matches[1])
	if convErr != nil {
		return 0, errors.Wrap(convErr, "failed to convert ID to integer")
	}
	return id, nil
}
