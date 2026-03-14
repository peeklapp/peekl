package models

type ApiClient interface {
	GetRootCA() (string, error)
	SubmitCertificateRequest(string, string) error
	RetrieveSignedCertificate(string, string) (string, error)
	GetCatalog(string) ([]Resource, []Role, []string, map[string]any, error)
	RetrieveFile(string, string, string) (string, error)
	RetrieveTemplate(string, string, string) (string, error)
}
