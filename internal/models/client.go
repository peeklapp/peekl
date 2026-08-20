package models

type ApiClient interface {
	GetRootCA() (string, error)
	SubmitCertificateRequest(string, string) error
	RetrieveSignedCertificate(string) (string, error)
	GetCatalog(string) (string, string, string, string, error)
	DownloadFile(string, string) error
}
