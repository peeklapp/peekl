package models

type ResourceRequire struct {
	Title string `yaml:"title" json:"title"`
	Type  string `yaml:"type" json:"type"`
}

type Resource struct {
	Title   string          `yaml:"title" json:"title"`
	Type    string          `yaml:"type" json:"type"`
	Data    any             `yaml:"data" json:"data"`
	Present bool            `yaml:"present" json:"present"`
	Require ResourceRequire `yaml:"require" json:"require"`
}

type ResourceResult struct {
	Created bool `yaml:"created" json:"created"`
	Updated bool `yaml:"updated" json:"updated"`
	Deleted bool `yaml:"deleted" json:"deleted"`
	Failed  bool `yaml:"failed" json:"failed"`
}

type LoadedResource interface {
	Process(*Context) (ResourceResult, error)
	String() string
}
