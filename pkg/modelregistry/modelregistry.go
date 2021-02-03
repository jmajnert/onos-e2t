// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package modelregistry

import (
	"fmt"
	types "github.com/onosproject/onos-api/go/onos/e2t/e2sm"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"plugin"
)

var log = logging.GetLogger("modelregistry")

type ModelType string
type ModelVersion string
type ModelFullName string

// ModelRegistry is the object for the saving information about device models
type ModelRegistry struct {
	ModelPlugins map[ModelFullName]ServiceModel
}

// ServiceModel is a set of methods that each model plugin should implement
type ServiceModel interface {
	ServiceModelData() (string, string, string)
	IndicationHeaderASN1toProto(asn1Bytes []byte) ([]byte, error)
	IndicationHeaderProtoToASN1(protoBytes []byte) ([]byte, error)
	IndicationMessageASN1toProto(asn1Bytes []byte) ([]byte, error)
	IndicationMessageProtoToASN1(protoBytes []byte) ([]byte, error)
	RanFuncDescriptionASN1toProto(asn1Bytes []byte) ([]byte, error)
	RanFuncDescriptionProtoToASN1(protoBytes []byte) ([]byte, error)
	EventTriggerDefinitionASN1toProto(asn1Bytes []byte) ([]byte, error)
	EventTriggerDefinitionProtoToASN1(protoBytes []byte) ([]byte, error)
	ActionDefinitionASN1toProto(asn1Bytes []byte) ([]byte, error)
	ActionDefinitionProtoToASN1(protoBytes []byte) ([]byte, error)
	DecodeRanFunctionDescription(asn1bytes []byte) (*types.RanfunctionNameDef, *types.RicEventTriggerList, *types.RicReportList, error)
}

// RegisterModelPlugin adds an external model plugin to the model registry at startup
// or through the 'admin' gRPC interface. Once plugins are loaded they cannot be unloaded
func (registry *ModelRegistry) RegisterModelPlugin(moduleName string) (ModelType, ModelVersion, error) {
	log.Info("Loading module ", moduleName)
	modelPluginModule, err := plugin.Open(moduleName)
	if err != nil {
		log.Warnf("Unable to load module %s %s", moduleName, err)
		return "", "", err
	}
	symbolMP, err := modelPluginModule.Lookup("ServiceModel")
	if err != nil {
		log.Warn("Unable to find ServiceModel in module ", moduleName, err)
		return "", "", err
	}
	serviceModelPlugin, ok := symbolMP.(ServiceModel)
	if !ok {
		log.Warnf("Unable to use ServiceModelPlugin in %s", moduleName)
		return "", "", fmt.Errorf("symbol loaded from module %s is not a ServiceModel",
			moduleName)
	}
	name, version, _ := serviceModelPlugin.ServiceModelData()
	log.Infof("Loaded %s %s from %s", name, version, moduleName)
	fullName := ToModelName(ModelType(name), ModelVersion(version))
	registry.ModelPlugins[fullName] = serviceModelPlugin

	return ModelType(name), ModelVersion(version), nil
}

// ToModelName simply joins together model type and version in a consistent way
func ToModelName(name ModelType, version ModelVersion) ModelFullName {
	return ModelFullName(fmt.Sprintf("%s-%s", name, version))
}