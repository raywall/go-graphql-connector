package models

import "encoding/json"

type ColaboradorConvenio struct {
	CodigoIdentificacaoPessoa string    `json:"codigoIdentificacaoPessoa"`
	CodigoConvenioContrato    string    `json:"codigoConvenioContrato"`
	Vinculos                  []Vinculo `json:"vinculos"`
}

type Vinculo struct {
	CodigoCons                   string                `json:"codigoCons"`
	CodigoMbf                    string                `json:"codigoMbf"`
	DomicilioBancario            DomicilioBancario     `json:"domicilioBancario"`
	DataBb                       string                `json:"dataBb"`
	DataCm                       string                `json:"dataCm"`
	DataDb                       string                `json:"dataDb"`
	DataExtc                     string                `json:"dataExtc"`
	DataNc                       string                `json:"dataNc"`
	DataTcB                      string                `json:"dataTcB"`
	DataTrl                      string                `json:"dataTrl"`
	DataUppc                     string                `json:"dataUppc"`
	EspecieBeneficio             EspecieBeneficio      `json:"especieBeneficio"`
	FormaPagamento               FormaPagamento        `json:"formaPagamento"`
	CodigoIdentificacaoInquilino string                `json:"codigoIdentificacaoInquilino"`
	CodigoIdentificacaoRl        string                `json:"codigoIdentificacaoRl"`
	IndicadorEb                  bool                  `json:"indicadorEb"`
	IndicadorEe                  bool                  `json:"indicadorEe"`
	IndicadorOr                  bool                  `json:"indicadorOr"`
	IndicadorPa                  FormaPagamento        `json:"indicadorPa"`
	IndicadorPep                 FormaPagamento        `json:"indicadorPep"`
	IndicadorPro                 bool                  `json:"indicadorPro"`
	IndicadorRl                  bool                  `json:"indicadorRl"`
	IndicadorRp                  bool                  `json:"indicadorRp"`
	IndicadorSben                FormaPagamento        `json:"indicadorSben"`
	IndicadorCj                  bool                  `json:"indicadorCj"`
	MarcaCon                     MarcaCon              `json:"marcaCon"`
	NomeDc                       string                `json:"nomeDc"`
	OrigemMc                     []OrigenMc            `json:"origemMc"`
	QuantidadeEmAs               int                   `json:"quantidadeEmAs"`
	QuantidadeEma                int                   `json:"quantidadeEma"`
	QuantidadeEmp                int                   `json:"quantidadeEmp"`
	QuantidadeEmr                int                   `json:"quantidadeEmr"`
	QuantidadeEms                int                   `json:"quantidadeEms"`
	QuantidadePl                 int                   `json:"quantidadePl"`
	SiglaUfp                     string                `json:"siglaUfp"`
	TipoBloq                     FormaPagamento        `json:"tipoBloq"`
	TipoVinCol                   string                `json:"tipoVinCol"`
	ValorCbene                   float64               `json:"valorCbene"`
	ValorLib                     float64               `json:"valorLib"`
	MensagemProcessamento        MensagemProcessamento `json:"mensagemProcessamento"`
}

type DomicilioBancario struct {
	CodigoAgencia string `json:"codigoAgencia"`
	CodigoBanco   int    `json:"codigoBanco"`
	NumeroConta   string `json:"numeroConta"`
}

type EspecieBeneficio struct {
	Codigo    int    `json:"codigo"`
	Descricao string `json:"descricao"`
}

type FormaPagamento struct {
	Codigo    int    `json:"codigo"`
	Descricao string `json:"descricao"`
}

type MarcaCon struct {
	ValorComp float64 `json:"valorComp"`
	ValorDae  float64 `json:"valorDae"`
	ValorDisC float64 `json:"valorDisC"`
	ValorDisE float64 `json:"valorDisE"`
	ValorLimC float64 `json:"valorLimC"`
	ValorLiq  float64 `json:"valorLiq"`
	ValorMcb  float64 `json:"valorMcb"`
}

type OrigenMc struct {
	Fonte           string `json:"fonte"`
	DataAtualizacao string `json:"dataAtualizacao"`
}

type MensagemProcessamento struct {
	Codigos   string `json:"codigos"`
	Descricao string `json:"descricao"`
}

func (c *ColaboradorConvenio) ToJSON() (string, error) {
	jsonData, err := json.Marshal(*c)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
