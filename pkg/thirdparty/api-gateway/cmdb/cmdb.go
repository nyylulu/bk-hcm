package cmdb

import (
	"hcm/pkg/cc"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	apigateway "hcm/pkg/thirdparty/api-gateway"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/thirdparty/esb/cmdb"
	"hcm/pkg/tools/ssl"

	"github.com/prometheus/client_golang/prometheus"
)

// NewClient initialize a new cmdbApiGateWay client
func NewClient(cfg *cc.ApiGateway, reg prometheus.Registerer, esbClient esb.Client) (cmdb.Client, error) {
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
			Name:    "cmdbApiGateWay",
			Servers: cfg.Endpoints,
		},
		MetricOpts: client.MetricOption{Register: reg},
	}
	restCli := rest.NewClient(c, "/api/v3")
	return &cmdbApiGateWay{
		config: cfg,
		client: restCli,
		Client: esbClient.Cmdb(),
	}, nil
}

// cmdbApiGateWay is an esb client to request cmdbApiGateWay.
type cmdbApiGateWay struct {
	config *cc.ApiGateway
	// http client instance
	client rest.ClientInterface
	// fall back to esbCall
	cmdb.Client
}

// ListBizHost ...
func (c *cmdbApiGateWay) ListBizHost(kt *kit.Kit, req *cmdb.ListBizHostParams) (
	*cmdb.ListBizHostResult, error) {

	return apigateway.ApiGatewayCall[cmdb.ListBizHostParams, cmdb.ListBizHostResult](c.client, c.config, rest.POST,
		kt, req, "/hosts/app/%d/list_hosts", req.BizID)
}

// 其他请求使用esb 接口
