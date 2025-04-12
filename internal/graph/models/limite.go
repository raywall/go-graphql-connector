package models

import "encoding/json"

type LimiteOperacional struct {
	PercentualMaximoMargemConsignavel            float64  `json:"percentualMaximoMargemConsignavel"`
	PercentualMaximoMargemConsignavelRestricao   *float64 `json:"percentualMaximoMargemConsignavelRestricao"`
	TipoMargemConsignavelRestricao               *string  `json:"tipoMargemConsignavelRestricao"`
	TipoTaxaMaximaRegulada                       string   `json:"tipoTaxaMaximaRegulada"`
	PercentualTaxaMaximaRegulada                 float64  `json:"percentualTaxaMaximaRegulada"`
	QuantidadeMinimaPrazo                        int      `json:"quantidadeMinimaPrazo"`
	QuantidadeMaximaPrazo                        int      `json:"quantidadeMaximaPrazo"`
	QuantidadeMaximaContratacoes                 int      `json:"quantidadeMaximaContratacoes"`
	ValorMinimoContratacao                       *float64 `json:"valorMinimoContratacao"`
	ValorMaximoContratacao                       float64  `json:"valorMaximoContratacao"`
	ValorMinimoParcela                           float64  `json:"valorMinimoParcela"`
	IdadeMinimaContratacao                       int      `json:"idadeMinimaContratacao"`
	IdadeMaximaContratacao                       int      `json:"idadeMaximaContratacao"`
	ValorMinimoTroco                             float64  `json:"valorMinimoTroco"`
	QuantidadeMaximaDiasAtraso                   int      `json:"quantidadeMaximaDiasAtraso"`
	PercentualMinimoSaldoDevedorQuitado          *float64 `json:"percentualMinimoSaldoDevedorQuitado"`
	QuantidadeMinimaParcelasPagas                int      `json:"quantidadeMinimaParcelasPagas"`
	QuantidadeMaximaParcelasAtraso               int      `json:"quantidadeMaximaParcelasAtraso"`
	QuantidadeMaximaContratosRefinanciadosPorVez int      `json:"quantidadeMaximaContratosRefinanciadosPorVez"`
	QuantidadeMinimaParcelasVencer               int      `json:"quantidadeMinimaParcelasVencer"`
}

func (l *LimiteOperacional) ToJSON() (string, error) {
	jsonData, err := json.Marshal(*l)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
