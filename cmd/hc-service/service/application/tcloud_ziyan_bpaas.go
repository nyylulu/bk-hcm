package application

import (
	cloudadaptor "hcm/cmd/hc-service/logics/cloud-adaptor"
	"hcm/cmd/hc-service/service/capability"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/client"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// InitApplicationService initial the application service
func InitApplicationService(cap *capability.Capability) {
	a := &application{
		ad: cap.CloudAdaptor,
		cs: cap.ClientSet,
	}

	h := rest.NewHandler()

	h.Add("QueryTCloudZiyanBPaasApplicationDetail", "POST", "/vendors/tcloud-ziyan/application/bpaas/query",
		a.QueryTCloudZiyanBPaasApplicationDetail)

	h.Load(cap.WebService)
}

type application struct {
	ad *cloudadaptor.CloudAdaptorClient
	cs *client.ClientSet
}

// QueryTCloudZiyanBPaasApplicationDetail ...
func (a *application) QueryTCloudZiyanBPaasApplicationDetail(cts *rest.Contexts) (any, error) {
	req := new(hcservice.GetBPaasApplicationReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ziyan, err := a.ad.TCloudZiyan(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}
	bpaasDetail, err := ziyan.GetBPaasApplicationDetail(cts.Kit, req.BPaasSN)
	if err != nil {
		logs.Errorf("fail to get bpaas application detail, err: %v, application id: %v, rid: %s",
			err, req.BPaasSN, cts.Kit.Rid)
		return nil, err
	}
	return bpaasDetail, nil
}
