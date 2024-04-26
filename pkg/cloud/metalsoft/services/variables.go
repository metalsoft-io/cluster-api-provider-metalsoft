package services

import (
	"github.com/metalsoft-io/cluster-api-provider-metalsoft/pkg/cloud/metalsoft"
	metalcloud "github.com/metalsoft-io/metal-cloud-sdk-go/v3"
)

type VariablesService struct {
	Client *metalsoft.MetalSoftClient
}

func NewVariablesService(client *metalsoft.MetalSoftClient) *VariablesService {
	return &VariablesService{
		Client: client,
	}
}

func (service *VariablesService) CreateVariable(variableName, variableValue string) (*metalcloud.Variable, error) {
	variable := metalcloud.Variable{
		VariableName: variableName,
		VariableJSON: variableValue,
	}
	createdVariable, err := service.Client.VariableCreate(variable)
	if err != nil {
		return nil, err
	}
	return createdVariable, nil
}

func (service *VariablesService) GetVariable(variableID int) (*metalcloud.Variable, error) {
	variable, err := service.Client.VariableGet(variableID)
	if err != nil {
		return nil, err
	}
	return variable, nil
}
