package models

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
)

type Resource struct {
	Type          string    `hcl:"type,label"`
	Title         string    `hcl:"title,label"`
	RawParameters cty.Value `hcl:"parameters,optional"`
	Parameters    map[string]any
	Present       *bool       `hcl:"present,optional"`
	When          string      `hcl:"when,optional"`
	Register      string      `hcl:"register,optional"`
	RawWith       []cty.Value `hcl:"with,optional"`
	With          []map[string]any
}

func (r *Resource) Resolve() error {
	parameters, err := CtyToMap(r.RawParameters)
	if err != nil {
		return fmt.Errorf("invalid parameters for resource '%s' : %w", r.Title, err)
	}
	r.Parameters = parameters

	r.With = make([]map[string]any, len(r.RawWith))
	for i, v := range r.RawWith {
		item, err := CtyToMap(v)
		if err != nil {
			return fmt.Errorf("invalid 'with' item %d for resource '%s' : %w", i, r.Title, err)
		}
		r.With[i] = item
	}

	return nil
}

type ResourceResult struct {
	Created bool `json:"created"`
	Updated bool `json:"updated"`
	Deleted bool `json:"deleted"`
	Failed  bool `json:"failed"`
}

type LoadedResource interface {
	Process(*ResourceContext) (ResourceResult, error)
	Validate() error
	When() string
	Register() string
	String() string
}
