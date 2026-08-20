package facts

import (
	"encoding/json"
	"github.com/peeklapp/peekl/internal/models"
)

func FactsToMap(facts models.Facts) (map[string]any, error) {
	var factsMap map[string]any
	jsonFacts, _ := json.Marshal(facts)
	if err := json.Unmarshal(jsonFacts, &factsMap); err != nil {
		return nil, err
	}
	return factsMap, nil
}
