package models

type NodeInventory struct {
	Name      string         `yaml:"name" json:"name" mapstructure:"name" hcl:"name"`
	Roles     []string       `yaml:"roles" json:"roles" mapstructure:"roles" hcl:"roles"`
	Resources []*Resource    `yaml:"resources" json:"resources" mapstructure:"resources" hcl:"resource,block"`
	Groups    []string       `yaml:"groups" json:"groups" mapstructure:"groups" hcl:"groups"`
	Tags      []string       `yaml:"tags" json:"tags" mapstructure:"tags" hcl:"tags"`
	Variables map[string]any `yaml:"variables" json:"variables" mapstructure:"variables"`
}

type GroupInventory struct {
	Name      string         `yaml:"name" json:"name" mapstructure:"name" hcl:"name"`
	Roles     []string       `yaml:"roles" json:"roles" mapstructure:"roles" hcl:"roles"`
	Resources []*Resource    `yaml:"resources" json:"resources" mapstructure:"resources" hcl:"resource,block"`
	Tags      []string       `yaml:"tags" json:"tags" mapstructure:"tags" hcl:"tags"`
	Variables map[string]any `yaml:"variables" json:"variables" mapstructure:"variables"`
}
