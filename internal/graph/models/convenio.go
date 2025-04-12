package models

import (
	"encoding/json"
)

type Convenio struct {
	CodigoConvenio                   int                      `json:"codigoConvenio"`
	NomeConvenio                     string                   `json:"nomeConvenio"`
	ApelidoConvenio                  string                   `json:"apelidoConvenio"`
	CodigoFq3PessoaJuridica          string                   `json:"codigoFq3PessoaJuridica"`
	CodigoTq3InquilinoPessoaJuridica string                   `json:"codigoTq3InquilinoPessoaJuridica"`
	GrupoConsignacao                 *string                  `json:"grupoConsignacao"`
	NumeroContratoMae                string                   `json:"numeroContratoMae"`
	InstituicaoFinanceira            string                   `json:"instituicaoFinanceira"`
	Segmento                         string                   `json:"segmento"`
	SituacaoConvenio                 string                   `json:"situacaoConvenio"`
	PossibilidadeHabilitacao         PossibilidadeHabilitacao `json:"possibilidadeHabilitacao"`
}

type PossibilidadeHabilitacao struct {
	Plataformas []string `json:"plataformas"`
	Modalidades []string `json:"modalidades"`
	Canais      []string `json:"canais"`
}

func (c *Convenio) ToJSON() (string, error) {
	jsonData, err := json.Marshal(*c)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
