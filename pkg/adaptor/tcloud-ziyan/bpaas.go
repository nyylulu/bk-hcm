package ziyan

import (
	"fmt"

	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	bpaas "hcm/pkg/thirdparty/tencentcloud/bpaas/v20181217"
)

// GetBPaasApplicationDetail 查询申请单详情
func (t *ZiyanAdpt) GetBPaasApplicationDetail(kt *kit.Kit, applicationID uint64) (
	*bpaas.GetBpaasApplicationDetailResponseParams, error) {

	if applicationID == 0 {
		return nil, errf.New(errf.InvalidParameter, "bpaas application can not be zero")
	}

	client, err := t.clientSet.BPaasClient()
	if err != nil {
		return nil, fmt.Errorf("init tcloud ziyan bpaas client failed, err: %v", err)
	}

	req := bpaas.NewGetBpaasApplicationDetailRequest()
	req.ApplicationId = &applicationID

	resp, err := client.GetBpaasApplicationDetailWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("get tcloud ziyan bpaas application detail failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp.Response, nil
}
