package models

type RawCatalog struct {
	GlobalResources []Resource
	Roles           []Role
	Facts           *Facts
	Tags            []string
	Variables       map[string]any
}
