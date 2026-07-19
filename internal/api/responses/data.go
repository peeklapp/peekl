package responses

type RetrieveFile struct {
	AgentCacheIsValid bool
	Filename          string
	Content           string
}

type RetrieveTemplate struct {
	AgentCacheIsValid bool
	TemplateName      string
	Content           string
}
