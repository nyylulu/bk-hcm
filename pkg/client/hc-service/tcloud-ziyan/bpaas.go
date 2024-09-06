package hcziyancli

import (
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// ApplicationClient is hc service bpaas api client.
type ApplicationClient struct {
	client rest.ClientInterface
}

// NewApplicationClient create a new account api client.
func NewApplicationClient(client rest.ClientInterface) *ApplicationClient {
	return &ApplicationClient{
		client: client,
	}
}

// QueryTCloudZiyanBPaasApplicationDetail 查询bpaas申请单详情
func (a *ApplicationClient) QueryTCloudZiyanBPaasApplicationDetail(kt *kit.Kit,
	req *hcservice.GetBPaasApplicationReq) (*any, error) {

	return common.Request[hcservice.GetBPaasApplicationReq, any](
		a.client, rest.POST, kt, req, "/application/bpaas/query")
}
