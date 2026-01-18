package responses

type RetrieveSignedCertificate struct {
	Certificate string `json:"certificate"`
}

type GetRootCA struct {
	Certificate string `json:"certificate"`
}
