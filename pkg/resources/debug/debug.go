package debug

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/mitchellh/mapstructure"
	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/resources"
	"github.com/sirupsen/logrus"
)

type DebugData struct {
	Message string `mapstructure:"message"`
}

type DebugResource struct {
	resources.CommonFieldResource
	Data DebugData
}

func (d *DebugResource) Process(context *models.ResourceContext) (models.ResourceResult, error) {
	var result models.ResourceResult

	jsonFacts, err := json.Marshal(context.Facts)
	if err != nil {
		result.Failed = true
		return result, err
	}
	var factsMap map[string]any
	err = json.Unmarshal(jsonFacts, &factsMap)
	if err != nil {
		result.Failed = true
		return result, err
	}

	variables := context.Variables
	variables["facts"] = factsMap

	tmpl, err := template.New(d.Title).Parse(d.Data.Message)
	if err != nil {
		result.Failed = true
		return result, err
	}

	var output bytes.Buffer
	err = tmpl.Execute(&output, variables)
	if err != nil {
		result.Failed = true
		return result, err
	}

	logrus.Info(output.String())

	return result, nil
}

func (d *DebugResource) String() string {
	return fmt.Sprintf("%s / '%s'", d.Type, d.Title)
}

func (d *DebugResource) When() string {
	return d.WhenCondition
}

func (d *DebugResource) Register() string {
	return d.RegisterVariable
}

func (d *DebugResource) Validate() error {
	validationErrors := []models.ValidationError{}

	if d.Data.Message == "" {
		validationErrors = append(
			validationErrors,
			models.ValidationError{
				FieldName:    "message",
				ViolatedRule: "Field cannot be empty",
			},
		)
	}

	_, err := template.New(d.Title).Parse(d.Data.Message)
	if err != nil {
		validationErrors = append(
			validationErrors,
			models.ValidationError{
				FieldName:    "message",
				ViolatedRule: fmt.Sprintf("Template is not valid : %s", err.Error()),
			},
		)
	}

	if len(validationErrors) > 0 {
		return models.ResourceValidationError{
			Type:             d.Type,
			Title:            d.Title,
			ValidationErrors: validationErrors,
		}
	}

	return nil
}

func NewDebugResource(resource *models.Resource, dataField any) (*DebugResource, error) {
	var debugResource DebugResource

	debugResource.Title = resource.Title
	debugResource.Type = resource.Type
	debugResource.Present = *resource.Present
	debugResource.WhenCondition = resource.When
	debugResource.RegisterVariable = resource.Register

	err := mapstructure.Decode(dataField, &debugResource.Data)
	if err != nil {
		return &debugResource, err
	}

	return &debugResource, nil
}
