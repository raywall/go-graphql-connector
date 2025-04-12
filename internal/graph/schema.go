package graph

import (
	"github.com/graphql-go/graphql"
)

func CreateSchema(res Resolver) (*graphql.Schema, error) {
	possibilidadeHabilitacaoType := graphql.NewObject(graphql.ObjectConfig{
		Name: "PossibilidadeHabilitacao",
		Fields: graphql.Fields{
			"plataformas": &graphql.Field{Type: graphql.NewList(graphql.String)},
			"modalidades": &graphql.Field{Type: graphql.NewList(graphql.String)},
			"canais":      &graphql.Field{Type: graphql.NewList(graphql.String)},
		},
	})

	convenioType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Convenio",
		Fields: graphql.Fields{
			"codigoConvenio":                   &graphql.Field{Type: graphql.Int},
			"nomeConvenio":                     &graphql.Field{Type: graphql.String},
			"apelidoConvenio":                  &graphql.Field{Type: graphql.String},
			"codigoFq3PessoaJuridica":          &graphql.Field{Type: graphql.String},
			"codigoTq3InquilinoPessoaJuridica": &graphql.Field{Type: graphql.String},
			"grupoConsignacao":                 &graphql.Field{Type: graphql.String},
			"numeroContratoMae":                &graphql.Field{Type: graphql.String},
			"instituicaoFinanceira":            &graphql.Field{Type: graphql.String},
			"segmento":                         &graphql.Field{Type: graphql.String},
			"situacaoConvenio":                 &graphql.Field{Type: graphql.String},
			"possibilidadeHabilitacao":         &graphql.Field{Type: possibilidadeHabilitacaoType},
		},
	})

	limiteOperacionalType := graphql.NewObject(graphql.ObjectConfig{
		Name: "LimiteOperacional",
		Fields: graphql.Fields{
			"percentualMaximoMargemConsignavel":            &graphql.Field{Type: graphql.Float},
			"percentualMaximoMargemConsignavelRestricao":   &graphql.Field{Type: graphql.Float},
			"tipoMargemConsignavelRestricao":               &graphql.Field{Type: graphql.String},
			"tipoTaxaMaximaRegulada":                       &graphql.Field{Type: graphql.String},
			"percentualTaxaMaximaRegulada":                 &graphql.Field{Type: graphql.Float},
			"quantidadeMinimaPrazo":                        &graphql.Field{Type: graphql.Int},
			"quantidadeMaximaPrazo":                        &graphql.Field{Type: graphql.Int},
			"quantidadeMaximaContratacoes":                 &graphql.Field{Type: graphql.Int},
			"valorMinimoContratacao":                       &graphql.Field{Type: graphql.Float},
			"valorMaximoContratacao":                       &graphql.Field{Type: graphql.Float},
			"valorMinimoParcela":                           &graphql.Field{Type: graphql.Float},
			"idadeMinimaContratacao":                       &graphql.Field{Type: graphql.Int},
			"idadeMaximaContratacao":                       &graphql.Field{Type: graphql.Int},
			"valorMinimoTroco":                             &graphql.Field{Type: graphql.Float},
			"quantidadeMaximaDiasAtraso":                   &graphql.Field{Type: graphql.Int},
			"percentualMinimoSaldoDevedorQuitado":          &graphql.Field{Type: graphql.Float},
			"quantidadeMinimaParcelasPagas":                &graphql.Field{Type: graphql.Int},
			"quantidadeMaximaParcelasAtraso":               &graphql.Field{Type: graphql.Int},
			"quantidadeMaximaContratosRefinanciadosPorVez": &graphql.Field{Type: graphql.Int},
			"quantidadeMinimaParcelasVencer":               &graphql.Field{Type: graphql.Int},
		},
	})

	prazoFundingType := graphql.NewObject(graphql.ObjectConfig{
		Name: "PrazoFunding",
		Fields: graphql.Fields{
			"prazo":       &graphql.Field{Type: graphql.Int},
			"taxaFunding": &graphql.Field{Type: graphql.Float},
		},
	})

	taxaFundingType := graphql.NewObject(graphql.ObjectConfig{
		Name: "TaxaFunding",
		Fields: graphql.Fields{
			"dataFunding": &graphql.Field{Type: graphql.String},
			"prazos":      &graphql.Field{Type: graphql.NewList(prazoFundingType)},
		},
	})

	domicilioBancarioType := graphql.NewObject(graphql.ObjectConfig{
		Name: "DomicilioBancario",
		Fields: graphql.Fields{
			"codigoAgencia": &graphql.Field{Type: graphql.String},
			"codigoBanco":   &graphql.Field{Type: graphql.Int},
			"numeroConta":   &graphql.Field{Type: graphql.String},
		},
	})

	especieBeneficioType := graphql.NewObject(graphql.ObjectConfig{
		Name: "EspecieBeneficio",
		Fields: graphql.Fields{
			"codigo":    &graphql.Field{Type: graphql.Int},
			"descricao": &graphql.Field{Type: graphql.String},
		},
	})

	formaPagamentoType := graphql.NewObject(graphql.ObjectConfig{
		Name: "FormaPagamento",
		Fields: graphql.Fields{
			"codigo":    &graphql.Field{Type: graphql.Int},
			"descricao": &graphql.Field{Type: graphql.String},
		},
	})

	marcaConType := graphql.NewObject(graphql.ObjectConfig{
		Name: "MarcaCon",
		Fields: graphql.Fields{
			"valorComp": &graphql.Field{Type: graphql.Float},
			"valorDae":  &graphql.Field{Type: graphql.Float},
			"valorDisC": &graphql.Field{Type: graphql.Float},
			"valorDisE": &graphql.Field{Type: graphql.Float},
			"valorLimC": &graphql.Field{Type: graphql.Float},
			"valorLiq":  &graphql.Field{Type: graphql.Float},
			"valorMcb":  &graphql.Field{Type: graphql.Float},
		},
	})

	origemMcType := graphql.NewObject(graphql.ObjectConfig{
		Name: "OrigemMc",
		Fields: graphql.Fields{
			"fonte":           &graphql.Field{Type: graphql.String},
			"dataAtualizacao": &graphql.Field{Type: graphql.String},
		},
	})

	mensagemProcessamentoType := graphql.NewObject(graphql.ObjectConfig{
		Name: "MensagemProcessamento",
		Fields: graphql.Fields{
			"codigos":   &graphql.Field{Type: graphql.String},
			"descricao": &graphql.Field{Type: graphql.String},
		},
	})

	vinculoType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Vinculo",
		Fields: graphql.Fields{
			"codigoCons":                   &graphql.Field{Type: graphql.String},
			"codigoMbf":                    &graphql.Field{Type: graphql.String},
			"domicilioBancario":            &graphql.Field{Type: domicilioBancarioType},
			"dataBb":                       &graphql.Field{Type: graphql.String},
			"dataCm":                       &graphql.Field{Type: graphql.String},
			"dataDb":                       &graphql.Field{Type: graphql.String},
			"dataExtc":                     &graphql.Field{Type: graphql.String},
			"dataNc":                       &graphql.Field{Type: graphql.String},
			"dataTcB":                      &graphql.Field{Type: graphql.String},
			"dataTrl":                      &graphql.Field{Type: graphql.String},
			"dataUppc":                     &graphql.Field{Type: graphql.String},
			"especieBeneficio":             &graphql.Field{Type: especieBeneficioType},
			"formaPagamento":               &graphql.Field{Type: formaPagamentoType},
			"codigoIdentificacaoInquilino": &graphql.Field{Type: graphql.String},
			"codigoIdentificacaoRt":        &graphql.Field{Type: graphql.String},
			"indicadorEb":                  &graphql.Field{Type: graphql.Boolean},
			"indicadorEe":                  &graphql.Field{Type: graphql.Boolean},
			"indicadorOr":                  &graphql.Field{Type: graphql.Boolean},
			"indicadorPa":                  &graphql.Field{Type: formaPagamentoType},
			"indicadorPep":                 &graphql.Field{Type: formaPagamentoType},
			"indicadorPro":                 &graphql.Field{Type: graphql.Boolean},
			"indicadorRt":                  &graphql.Field{Type: graphql.Boolean},
			"indicadorRp":                  &graphql.Field{Type: graphql.Boolean},
			"indicadorSben":                &graphql.Field{Type: formaPagamentoType},
			"indicadorCj":                  &graphql.Field{Type: graphql.Boolean},
			"marcaCon":                     &graphql.Field{Type: marcaConType},
			"nomeDc":                       &graphql.Field{Type: graphql.String},
			"origemMc":                     &graphql.Field{Type: graphql.NewList(origemMcType)},
			"quantidadeEmAs":               &graphql.Field{Type: graphql.Int},
			"quantidadeEma":                &graphql.Field{Type: graphql.Int},
			"quantidadeEmp":                &graphql.Field{Type: graphql.Int},
			"quantidadeEmr":                &graphql.Field{Type: graphql.Int},
			"quantidadeEms":                &graphql.Field{Type: graphql.Int},
			"quantidadePl":                 &graphql.Field{Type: graphql.Int},
			"siglaUfp":                     &graphql.Field{Type: graphql.String},
			"tipoBloq":                     &graphql.Field{Type: formaPagamentoType},
			"tipoVinCol":                   &graphql.Field{Type: graphql.String},
			"valorCbene":                   &graphql.Field{Type: graphql.Float},
			"valorLib":                     &graphql.Field{Type: graphql.Float},
			"mensagemProcessamento":        &graphql.Field{Type: mensagemProcessamentoType},
		},
	})

	colaboradorConvenioType := graphql.NewObject(graphql.ObjectConfig{
		Name: "ColaboradorConvenio",
		Fields: graphql.Fields{
			"codigoIdentificacaoPessoa": &graphql.Field{Type: graphql.String},
			"codigoConvenioContrato":    &graphql.Field{Type: graphql.String},
			"vinculos":                  &graphql.Field{Type: graphql.NewList(vinculoType)},
		},
	})

	combinedDataType := graphql.NewObject(graphql.ObjectConfig{
		Name: "CombinedData",
		Fields: graphql.Fields{
			"convenio":            &graphql.Field{Type: convenioType},
			"limiteOperacional":   &graphql.Field{Type: limiteOperacionalType},
			"taxaFunding":         &graphql.Field{Type: taxaFundingType},
			"colaboradorConvenio": &graphql.Field{Type: colaboradorConvenioType},
		},
	})

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"dataSources": &graphql.Field{
				Type: combinedDataType,
				Args: graphql.FieldConfigArgument{
					"codigoConvenio": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: res.ResolveDataSource,
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
	})
	if err != nil {
		panic(err)
	}

	return &schema, err
}
