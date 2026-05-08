package requests

type SubmitCertificateRequest struct {
	NodeName string `json:"node_name" validate:"required"`
	CSR      string `json:"csr" validate:"required"`
}

type RetrieveSignedCertificate struct {
	CsrSignature string `json:"csr_signature" validate:"required"`
}
