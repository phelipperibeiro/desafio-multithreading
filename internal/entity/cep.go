package entity

import (
	"errors"
)

type Cep struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
}

func (c *Cep) Validate() error {
	if c.Cep == "" {
		return errors.New("cep is required")
	}
	// Add more validation logic if needed
	return nil
}
