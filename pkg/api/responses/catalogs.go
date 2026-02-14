package responses

import "github.com/peeklapp/peekl/pkg/models"

type GetCatalog struct {
	GlobalResource []models.Resource `json:"resources"`
	Roles          []models.Role     `json:"roles"`
	Tags           []string          `json:"tags"`
	Variables      map[string]any    `json:"variables"`
}
