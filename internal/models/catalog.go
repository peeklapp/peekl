package models

type RawCatalog struct {
	Roles       []Role
	Facts       *Facts
	Tags        []string
	Variables   map[string]any
	Environment string
	ApiClient   ApiClient
	CodePath    string
}
