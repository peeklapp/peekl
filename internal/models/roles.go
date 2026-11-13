package models

type RoleIncludeEntry struct {
	Name string `hcl:"name,label"`
	Path string `hcl:"path"`
	When string `hcl:"when,optional"`
}

type RoleInclude struct {
	Resources       []Resource `hcl:"resource,block"`
	LoadedResources []LoadedResource
	When            string
}

type RoleMeta struct {
	Requires []string `hcl:"requires,optional"`
}

type RoleEntrypoint struct {
	Resources       []Resource `hcl:"resource,block"`
	LoadedResources []LoadedResource
	Includes        []RoleIncludeEntry `hcl:"include,block"`
	Meta            *RoleMeta          `hcl:"meta,block"`
}

type Role struct {
	Name       string
	Entrypoint RoleEntrypoint
	Included   map[string]RoleInclude
	Variables  map[string]any
}
