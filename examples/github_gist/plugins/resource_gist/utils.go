package tf

import (
	"encoding/json"
	"fmt"
)

func tfAttributesToModel(s string) (*Attributes, error) {
	rd := Attributes{}
	err := json.Unmarshal([]byte(s), &rd)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal JSON resource data: %w", err)
	}

	err = rd.validate()
	if err != nil {
		return nil, fmt.Errorf("invalid resource data: %w", err)
	}

	return &rd, nil
}
