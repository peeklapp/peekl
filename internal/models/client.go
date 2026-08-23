package models

type ApiClient interface {
	EnrollAgent(string, string) (string, string, error)
	GetCatalog(string) (string, string, string, string, error)
	DownloadFile(string, string) error
}
