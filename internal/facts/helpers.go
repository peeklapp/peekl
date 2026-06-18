package facts

import (
	"encoding/json"
	"github.com/peeklapp/peekl/internal/models"
)

func FactsToMap(facts models.Facts) map[string]any {
	var factsMap map[string]any
	jsonFacts, _ := json.Marshal(facts)
	json.Unmarshal(jsonFacts, &factsMap)
	return factsMap
}
