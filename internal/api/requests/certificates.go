package requests

type SubmitCertificateRequest struct {
	CSR string `json:"csr" validate:"required"`
}

type RetrieveSignedCertificate struct {
	CsrSignature string `json:"csr_signature" validate:"required"`
}

type EnrollAgent struct {
	CSR   string `json:"csr" validate:"required"`
	Token string `json:"token" validate:"required"`
}
