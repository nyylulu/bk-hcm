package ziyan

import (
	"fmt"
	"strings"

	"hcm/pkg/adaptor/types"
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

	if ms := tryGetMultiSecret(client); ms != nil {
		return getBpaasApplicationDetailByMultiSec(kt, client, ms, applicationID)
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

func tryGetMultiSecret(client *bpaas.Client) *types.MultiSecret {
	if client.GetCredential() == nil {
		return nil
	}
	ms, ok := client.GetCredential().(*types.MultiSecret)
	if !ok {
		return nil
	}
	return ms
}

func getBpaasApplicationDetailByMultiSec(kt *kit.Kit, client *bpaas.Client, ms *types.MultiSecret,
	applicationID uint64) (*bpaas.GetBpaasApplicationDetailResponseParams, error) {

	req := bpaas.NewGetBpaasApplicationDetailRequest()
	req.ApplicationId = &applicationID
	var noPermErr error
	// 开启多秘钥的情况下，尝试使用每一个秘钥进行遍历
	for _, sec := range ms.GetSecrets() {
		client.WithSecretId(sec.CloudSecretID, sec.CloudSecretKey)
		resp, err := client.GetBpaasApplicationDetailWithContext(kt.Ctx, req)
		if err != nil {
			if strings.Contains(err.Error(), "Code=UnauthorizedOperation.PermissionDenied") {
				// 无权限
				logs.V(4).Infof("unauthorized to get ziyan bpaas application detail, err: %v, secID:%s, rid: %s",
					err, sec.CloudSecretID, kt.Rid)
				noPermErr = err
				continue
			}
			// 其他错误直接返回
			logs.Errorf("get tcloud ziyan bpaas application detail failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		return resp.Response, nil
	}
	// 所有秘钥都没有权限，返回其中一个无权限错误
	logs.Errorf("all secrets unauthorized to get ziyan bpaas application detail, lastErr: %s, rid: %s",
		noPermErr, kt.Rid)
	return nil, noPermErr
}
