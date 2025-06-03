package securitygroup

import (
	"fmt"
	"strings"

	cloudserver "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/converter"
	"hcm/pkg/ziyan"
)

func (svc *securityGroupSvc) batchAssociateTCloudZiyanCvms(cts *rest.Contexts,
	req *hcproto.SecurityGroupBatchAssociateCvmReq) (any, error) {

	err := svc.client.HCService().TCloudZiyan.SecurityGroup.BatchAssociateCvm(cts.Kit, req.SecurityGroupID,
		req.CvmIDs)
	if err != nil {
		logs.Errorf("fail to call hc service associate cloud cvm, err: %v, sg_id: %s, cloud_cvm_ids: %v, rid:%s",
			err, req.SecurityGroupID, req.CvmIDs, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

func (svc *securityGroupSvc) tcloudZiyanCloneSecurityGroup(kt *kit.Kit, bizID int64, sg *cloud.BaseSecurityGroup,
	req *cloudserver.SecurityGroupCloneReq) (*core.CreateResult, error) {

	// 打业务标签
	meta, err := ziyan.GetResourceMetaByBizForUser(kt, cmdb.CmdbClient(), bizID, req.Manager, req.BakManager)
	if err != nil {
		logs.Errorf("get resource meta by biz failed, err: %v, biz: %d, rid: %s", err, bizID, kt.Rid)
		return nil, err
	}
	cloneReq := &hcproto.TCloudSecurityGroupCloneReq{
		SecurityGroupID: sg.ID,
		Manager:         req.Manager,
		BakManager:      req.BakManager,
		ManagementBizID: bizID,
		TargetRegion:    req.TargetRegion,
		Tags:            meta.GetTagPairs(),
	}
	if req.Name == nil {
		cloneReq.GroupName = fmt.Sprintf("%s-copy", sg.Name)
	} else {
		cloneReq.GroupName = converter.PtrToVal(req.Name)
	}
	if err = validateZiyanSGName(cloneReq.GroupName); err != nil {
		logs.Errorf("validate security group name failed, err: %v, sg_name: %s, rid: %s", err, cloneReq.GroupName, kt.Rid)
		return nil, err
	}
	result, err := svc.client.HCService().TCloudZiyan.SecurityGroup.CloneSecurityGroup(kt, cloneReq)
	if err != nil {
		logs.Errorf("clone security group failed, err: %v, req: %+v, rid: %s", err, cloneReq, kt.Rid)
		return nil, err
	}
	return result, nil
}

const (
	// InvalidSGNameYunti 不合法的安全组名称, 安全组名称中不能包含的子串
	InvalidSGNameYunti = "云梯默认安全组"
)

func validateZiyanSGName(name string) error {
	if strings.Contains(name, InvalidSGNameYunti) {
		return errf.New(errf.InvalidParameter, fmt.Sprintf("name can not contain %s", InvalidSGNameYunti))
	}
	return nil
}
