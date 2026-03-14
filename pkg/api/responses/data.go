package responses

type RetrieveFile struct {
	Filename string
	Content  string
}

type RetrieveTemplate struct {
	TemplateName string
	Content      string
}
