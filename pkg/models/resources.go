package models

import (
	"encoding/json"
	"fmt"
)

type ValidationError struct {
	FieldName    string `json:"field_name"`
	ViolatedRule string `json:"violated_rule"`
}

type ResourceValidationError struct {
	Type             string
	Title            string
	ValidationErrors []ValidationError
}

func (r ResourceValidationError) Error() string {
	var outputString = fmt.Sprintf(
		"Invalid resource [%s / '%s'] : ",
		r.Type,
		r.Title,
	)
	jsonErrors, _ := json.Marshal(r.ValidationErrors)
	outputString += string(jsonErrors)
	return outputString
}

type ResourceRequire struct {
	Title string `yaml:"title" json:"title"`
	Type  string `yaml:"type" json:"type"`
}

type Resource struct {
	Title    string           `yaml:"title" json:"title"`
	Type     string           `yaml:"type" json:"type"`
	Data     map[string]any   `yaml:"data" json:"data"`
	Present  *bool            `yaml:"present" json:"present"`
	Require  ResourceRequire  `yaml:"require" json:"require"`
	When     string           `yaml:"when" json:"when"`
	Register string           `yaml:"register" json:"register"`
	With     []map[string]any `yaml:"with" json:"with"`
}

type ResourceResult struct {
	Created bool `yaml:"created" json:"created"`
	Updated bool `yaml:"updated" json:"updated"`
	Deleted bool `yaml:"deleted" json:"deleted"`
	Failed  bool `yaml:"failed" json:"failed"`
}

type LoadedResource interface {
	Process(*ResourceContext) (ResourceResult, error)
	Validate() error
	When() string
	Register() string
	String() string
}
