package fixture

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"weddinglive/internal/domain"
)

//go:embed demo.json
var demoData []byte

func Demo() (domain.State, error) {
	var state domain.State
	if err := json.Unmarshal(demoData, &state); err != nil {
		return domain.State{}, fmt.Errorf("decode embedded fixture: %w", err)
	}
	return state, nil
}
