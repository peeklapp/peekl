package models

type ResourceContext struct {
	Facts     *Facts
	Tags      []string
	Variables map[string]any
}

type RoleContext struct {
	Templates map[string]string
}
