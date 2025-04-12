package models

import "encoding/json"

type TaxaFunding struct {
	DataFunding string         `json:"dataFunding"`
	Prazos      []PrazoFunding `json:"prazos"`
}

type PrazoFunding struct {
	Prazo       int     `json:"prazo"`
	TaxaFunding float64 `json:"taxaFunding"`
}

func (f *TaxaFunding) ToJSON() (string, error) {
	jsonData, err := json.Marshal(*f)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
