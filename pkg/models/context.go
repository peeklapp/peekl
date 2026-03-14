package models

type ResourceContext struct {
	Facts       *Facts
	Tags        []string
	Variables   map[string]any
	Environment string
	ApiClient   ApiClient
}

type RoleContext struct {
	RoleName string
}
