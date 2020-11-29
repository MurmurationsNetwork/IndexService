package service

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/MurmurationsNetwork/MurmurationsServices/common/cryptoutil"
	"github.com/MurmurationsNetwork/MurmurationsServices/common/dateutil"
	"github.com/MurmurationsNetwork/MurmurationsServices/common/event"
	"github.com/MurmurationsNetwork/MurmurationsServices/common/httputil"
	"github.com/MurmurationsNetwork/MurmurationsServices/services/validation/internal/adapter/nats"
	"github.com/MurmurationsNetwork/MurmurationsServices/services/validation/internal/domain/node"
	"github.com/xeipuuv/gojsonschema"
)

var (
	ValidationService validationServiceInterface = &validationService{}
)

type validationServiceInterface interface {
	ValidateNode(node *node.Node)
}

type validationService struct{}

func (v *validationService) ValidateNode(node *node.Node) {
	document := gojsonschema.NewReferenceLoader(node.ProfileURL)
	data, err := document.LoadJSON()
	if err != nil {
		sendNodeValidationFailedEvent(node, []string{"Could not read from profile_url: " + node.ProfileURL})
		return
	}

	// Validate against the default schema.
	FailureReasons := validateAgainstSchemas([]string{"default-v1"}, document)
	if len(FailureReasons) != 0 {
		sendNodeValidationFailedEvent(node, FailureReasons)
		return
	}

	linkedSchemas, ok := getLinkedSchemas(data)
	if !ok {
		sendNodeValidationFailedEvent(node, []string{"Could not read linked_schemas from profile_url: " + node.ProfileURL})
		return
	}

	// Validate against schemes specify inside the profile data.
	FailureReasons = validateAgainstSchemas(linkedSchemas, document)
	if len(FailureReasons) != 0 {
		sendNodeValidationFailedEvent(node, FailureReasons)
		return
	}

	jsonStr, err := getJSONStr(node.ProfileURL)
	if err != nil {
		sendNodeValidationFailedEvent(node, []string{"Could not get JSON string from profile_url: " + node.ProfileURL})
		return
	}

	event.NewNodeValidatedPublisher(nats.Client()).Publish(event.NodeValidatedData{
		ProfileURL:    node.ProfileURL,
		ProfileHash:   cryptoutil.GetSHA256(jsonStr),
		ProfileStr:    jsonStr,
		LastValidated: dateutil.GetNowUnix(),
		Version:       node.Version,
	})
}

func getLinkedSchemas(data interface{}) ([]string, bool) {
	json, ok := data.(map[string]interface{})
	if !ok {
		return nil, false
	}
	_, ok = json["linked_schemas"]
	if !ok {
		return nil, false
	}
	arrInterface, ok := json["linked_schemas"].([]interface{})
	if !ok {
		return nil, false
	}

	var linkedSchemas = make([]string, 0)

	for _, data := range arrInterface {
		linkedSchema, ok := data.(string)
		if !ok {
			return nil, false
		}
		linkedSchemas = append(linkedSchemas, linkedSchema)
	}

	return linkedSchemas, true
}

func validateAgainstSchemas(linkedSchemas []string, document gojsonschema.JSONLoader) []string {
	FailureReasons := []string{}

	for _, linkedSchema := range linkedSchemas {
		schemaURL := os.Getenv("SCHEMAS_URL") + "/" + linkedSchema + ".json"

		schema, err := gojsonschema.NewSchema(gojsonschema.NewReferenceLoader(schemaURL))
		if err != nil {
			FailureReasons = append(FailureReasons, "Could not read from schema: "+schemaURL)
			continue
		}

		result, err := schema.Validate(document)
		if err != nil {
			FailureReasons = append(FailureReasons, "Error when trying to validate document: ", err.Error())
			continue
		}

		if !result.Valid() {
			FailureReasons = append(FailureReasons, parseValidateError(linkedSchema, result.Errors())...)
		}
	}

	return FailureReasons
}

func parseValidateError(schema string, resultErrors []gojsonschema.ResultError) []string {
	FailureReasons := make([]string, 0)
	for _, desc := range resultErrors {
		// Output string: "demo-v1.(root): url is required"
		FailureReasons = append(FailureReasons, schema+"."+desc.String())
	}
	return FailureReasons
}

func getJSONStr(source string) (string, error) {
	jsonByte, err := httputil.GetByte(source)
	if err != nil {
		return "", err
	}
	buffer := bytes.Buffer{}
	err = json.Compact(&buffer, jsonByte)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func sendNodeValidationFailedEvent(node *node.Node, FailureReasons []string) {
	event.NewNodeValidationFailedPublisher(nats.Client()).Publish(event.NodeValidationFailedData{
		ProfileURL:     node.ProfileURL,
		FailureReasons: FailureReasons,
		Version:        node.Version,
	})
}
