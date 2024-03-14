package services

import (
	"github.com/metalsoft-io/cluster-api-provider-metalsoft/pkg/cloud/metalsoft"
	metalcloud "github.com/metalsoft-io/metal-cloud-sdk-go/v2"
)

type VariablesService struct {
	*metalsoft.MetalSoftClient
}

func (service *VariablesService) createVariable(variableName, variableValue string) (*metalcloud.Variable, error) {
	variable := metalcloud.Variable{
		VariableName: variableName,
		VariableJSON: variableValue,
	}
	createdVariable, err := service.VariableCreate(variable)
	if err != nil {
		return nil, err
	}
	return createdVariable, nil
}

func (service *VariablesService) getVariable(variableID int) (*metalcloud.Variable, error) {
	variable, err := service.VariableGet(variableID)
	if err != nil {
		return nil, err
	}
	return variable, nil
}
