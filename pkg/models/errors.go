package models

import "fmt"

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

type EnvironmentNotFoundError struct {
	Environment string
}

func (e EnvironmentNotFoundError) Error() string {
	return fmt.Sprintf("The environment %s could not be found in the code folder", e.Environment)
}
