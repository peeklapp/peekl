package requests

type RetrieveFile struct {
	RoleName       string `json:"role_name"`
	Environment    string `json:"environment"`
	Filename       string `json:"filename"`
	AgentCacheHash string `json:"hash"`
}

type RetrieveTemplate struct {
	RoleName       string `json:"role_name"`
	Environment    string `json:"environment"`
	TemplateName   string `json:"template_name"`
	AgentCacheHash string `json:"hash"`
}
