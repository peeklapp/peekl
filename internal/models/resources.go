package models

import "github.com/zclconf/go-cty/cty"

type ResourceRequire struct {
	Title string `yaml:"title" json:"title"`
	Type  string `yaml:"type" json:"type"`
}

type Resource struct {
	Type     string                `yaml:"type" json:"type" hcl:"type,label"`
	Title    string                `yaml:"title" json:"title" hcl:"title,label"`
	Data     map[string]cty.Type   `yaml:"data" json:"data" hcl:"parameters"`
	Present  *bool                 `yaml:"present" json:"present" hcl:"present"`
	Require  ResourceRequire       `yaml:"require" json:"require" hcl:"require"`
	When     string                `yaml:"when" json:"when" hcl:"when"`
	Register string                `yaml:"register" json:"register" hcl:"register"`
	With     []map[string]cty.Type `yaml:"with" json:"with" hcl:"with"`
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
