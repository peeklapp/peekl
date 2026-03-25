package models

type IncludedResources struct {
	Resources       []Resource       `json:"resources" yaml:"resources" mapstructure:"resources"`
	LoadedResources []LoadedResource `json:"loaded_resources,omitempty"`
}

type Role struct {
	Name              string                       `json:"name" yaml:"string"`
	Resources         []Resource                   `json:"resources" yaml:"resources" mapstructure:"resources"`
	LoadedResources   []LoadedResource             `json:"loaded_resources,omitempty"`
	IncludedResources map[string]IncludedResources `json:"included_resources" yaml:"included_resources" mapstructure:"included_resources"`
	Variables         map[string]any               `json:"variables" yaml:"variables"`
}

type IncludeEntry struct {
	Name string `json:"name" yaml:"name" mapstructure:"name"`
}

type RoleMain struct {
	Resources      []Resource       `json:"resources" yaml:"resources" mapstructure:"resources"`
	LoadedResource []LoadedResource `json:"loaded_resources,omitempty"`
	Includes       []IncludeEntry   `json:"includes" yaml:"includes" mapstructure:"includes"`
}
