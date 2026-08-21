package models

import (
	"encoding/json"
	"fmt"
)

type RoleNotFoundError struct {
	RoleName string
}

func (e RoleNotFoundError) Error() string {
	return fmt.Sprintf("The role %s could not be found in the roles folder", e.RoleName)
}

type NodeNotFoundError struct {
	NodeName string
}

func (e NodeNotFoundError) Error() string {
	return fmt.Sprintf("The node %s could not be found in the inventory", e.NodeName)
}

type GroupNotFoundError struct {
	GroupName string
}

func (e GroupNotFoundError) Error() string {
	return fmt.Sprintf("The group %s could not be found in the inventory", e.GroupName)
}

type PendingCertificateNotFound struct {
	NodeName string
}

func (c PendingCertificateNotFound) Error() string {
	return fmt.Sprintf("No pending certificate found for node %s", c.NodeName)
}

type SignedCertificateNotFound struct {
	CsrSignature string
}

func (c SignedCertificateNotFound) Error() string {
	return fmt.Sprintf("No signed certificate found for node with signature %s", c.CsrSignature)
}

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

type ConfigurationValidationError struct {
	Path             string
	ValidationErrors []ValidationError
}

func (c ConfigurationValidationError) Error() string {
	var outputString = fmt.Sprintf(
		"Invalid configuration [%s] : ",
		c.Path,
	)
	jsonErrors, _ := json.Marshal(c.ValidationErrors)
	outputString += string(jsonErrors)
	return outputString
}
