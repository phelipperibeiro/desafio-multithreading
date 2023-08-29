package entity

import (
	"errors"
)

type Cep struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
}

func NewCep(cep string, logradouro string, complemento string, bairro string, localidade string, uf string) *Cep {
	return &Cep{
		Cep:         cep,
		Logradouro:  logradouro,
		Complemento: complemento,
		Bairro:      bairro,
		Localidade:  localidade,
		Uf:          uf,
	}
}

func (c *Cep) Validate() error {
	if c.Cep == "" {
		return errors.New("cep is required")
	}
	// Add more validation logic if needed
	return nil
}
