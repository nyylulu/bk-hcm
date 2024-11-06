package finops

import (
	"hcm/pkg/api/core"
	"hcm/pkg/cc"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	apigateway "hcm/pkg/thirdparty/api-gateway"
	"hcm/pkg/tools/ssl"

	"github.com/prometheus/client_golang/prometheus"
)

// OpProductBgIEG BG ID for IEG
const OpProductBgIEG = 4

// ListOpProductParam ...
type ListOpProductParam struct {
	// 筛选事业群id列表，不传筛选全部
	BgIds []int64 `json:"bg_ids"`
	// 要查询部门 id 列表(不传则筛选全部)
	DeptIds []int64 `json:"dept_ids"`
	// 要查询运营产品 id 列表(不传则筛选全部)
	OpProductIds []int64 `json:"op_product_ids"`
	// 要查询运营产品名称(支持全模糊匹配，不传则筛选全部)
	OpProductNames string        `json:"op_product_name"`
	Page           core.BasePage `json:"page"`
}

// OperationProduct 运营产品
type OperationProduct struct {
	BgId          int64  `json:"bg_id"`
	BgName        string `json:"bg_name"`
	BgShortName   string `json:"bg_short_name"`
	DeptId        int64  `json:"dept_id"`
	DeptName      string `json:"dept_name"`
	PlProductId   int64  `json:"pl_product_id"`
	PlProductName string `json:"pl_product_name"`
	OpProductId   int64  `json:"op_product_id"`
	OpProductName string `json:"op_product_name"`
	PrincipalName string `json:"principal_name"`
}

// ListOpProductResult ...
type ListOpProductResult struct {
	Count uint64             `json:"count"`
	Items []OperationProduct `json:"items"`
}

// Client FinOps Client
type Client interface {
	// ListOpProduct 查询全内部事业群运营产品
	ListOpProduct(kt *kit.Kit, params *ListOpProductParam) (*ListOpProductResult, error)
}

// NewClient initialize a new FinOps client
func NewClient(cfg *cc.ApiGateway, reg prometheus.Registerer) (Client, error) {
	tls := &ssl.TLSConfig{
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
		CertFile:           cfg.TLS.CertFile,
		KeyFile:            cfg.TLS.KeyFile,
		CAFile:             cfg.TLS.CAFile,
		Password:           cfg.TLS.Password,
	}
	cli, err := client.NewClient(tls)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &apigateway.Discovery{
			Name:    "fineOps",
			Servers: cfg.Endpoints,
		},
		MetricOpts: client.MetricOption{Register: reg},
	}
	restCli := rest.NewClient(c, "/api/v1")
	return &finOps{
		config: cfg,
		client: restCli,
	}, nil
}

// fineOps is an esb client to request fineOps.
type finOps struct {
	config *cc.ApiGateway
	// http client instance
	client rest.ClientInterface
}

// ListOpProduct 查询全内部事业群运营产品 get_op_product_meta
func (c *finOps) ListOpProduct(kt *kit.Kit, params *ListOpProductParam) (*ListOpProductResult, error) {

	return apigateway.ApiGatewayCall[ListOpProductParam, ListOpProductResult](c.client, c.config, rest.POST,
		kt, params, "/analysis/dm/meta/get/op_product/info")
}

// 其他请求使用esb 接口
