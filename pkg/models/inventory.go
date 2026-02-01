package models

type NodeInventory struct {
	Name      string         `yaml:"name" json:"name" mapstructure:"name"`
	Roles     []string       `yaml:"roles" json:"roles" mapstructure:"roles"`
	Resources []*Resource    `yaml:"resources" json:"resources" mapstructure:"resources"`
	Groups    []string       `yaml:"groups" json:"groups" mapstructure:"groups"`
	Tags      []string       `yaml:"tags" json:"tags" mapstructure:"tags"`
	Variables map[string]any `yaml:"variables" json:"variables" mapstructure:"variables"`
}

type GroupInventory struct {
	Name      string         `yaml:"name" json:"name" mapstructure:"name"`
	Roles     []string       `yaml:"roles" json:"roles" mapstructure:"roles"`
	Resources []*Resource    `yaml:"resources" json:"resources" mapstructure:"resources"`
	Tags      []string       `yaml:"tags" json:"tags" mapstructure:"tags"`
	Variables map[string]any `yaml:"variables" json:"variables" mapstructure:"variables"`
}
