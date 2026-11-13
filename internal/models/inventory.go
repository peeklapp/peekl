package models

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
)

type RoleInventory struct {
	Name         string    `hcl:"name,label"`
	RawVariables cty.Value `hcl:"variables,optional"`
	Variables    map[string]any
}

func (r *RoleInventory) Resolve() error {
	variables, err := CtyToMap(r.RawVariables)
	if err != nil {
		return fmt.Errorf("invalid variables for role '%s' : %w", r.Name, err)
	}
	r.Variables = variables
	return nil
}

func resolveRoles(roles []RoleInventory) error {
	for i := range roles {
		if err := roles[i].Resolve(); err != nil {
			return err
		}
	}
	return nil
}

type NodeInventoryFile struct {
	Node NodeInventory `hcl:"node,block"`
}

type NodeInventory struct {
	// SourcePath is the file the node was loaded from, relative to the code root.
	SourcePath   string
	Name         string          `hcl:"name,label"`
	Groups       []string        `hcl:"groups,optional"`
	Tags         []string        `hcl:"tags,optional"`
	Roles        []RoleInventory `hcl:"role,block"`
	RawVariables cty.Value       `hcl:"variables,optional"`
	Variables    map[string]any
}

func (n *NodeInventory) GetName() string {
	return n.Name
}

func (n *NodeInventory) SetSourcePath(path string) {
	n.SourcePath = path
}

func (n *NodeInventory) Resolve() error {
	variables, err := CtyToMap(n.RawVariables)
	if err != nil {
		return fmt.Errorf("invalid variables for node '%s' : %w", n.Name, err)
	}
	n.Variables = variables
	return resolveRoles(n.Roles)
}

type GroupInventoryFile struct {
	Group GroupInventory `hcl:"group,block"`
}

type GroupInventory struct {
	// SourcePath is the file the group was loaded from, relative to the code root.
	SourcePath   string
	Name         string          `hcl:"name,label"`
	Tags         []string        `hcl:"tags,optional"`
	Roles        []RoleInventory `hcl:"role,block"`
	RawVariables cty.Value       `hcl:"variables,optional"`
	Variables    map[string]any
}

func (g *GroupInventory) GetName() string {
	return g.Name
}

func (g *GroupInventory) SetSourcePath(path string) {
	g.SourcePath = path
}

func (g *GroupInventory) Resolve() error {
	variables, err := CtyToMap(g.RawVariables)
	if err != nil {
		return fmt.Errorf("invalid variables for group '%s' : %w", g.Name, err)
	}
	g.Variables = variables
	return resolveRoles(g.Roles)
}
