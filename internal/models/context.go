package models

type ResourceContext struct {
	CodePath    string
	Facts       *Facts
	Tags        []string
	Variables   map[string]any
	Environment string
	ApiClient   ApiClient
}

type RoleContext struct {
	RoleName string
}
