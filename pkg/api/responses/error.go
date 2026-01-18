package responses

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details"`
}
