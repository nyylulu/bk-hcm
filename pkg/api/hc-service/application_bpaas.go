package hcservice

import "hcm/pkg/criteria/validator"

// GetBPaasApplicationReq 查询bpaas申请单详情
type GetBPaasApplicationReq struct {
	BPaasSN   uint64 `json:"bpaas_sn"  validate:"required,gt=0"`
	AccountID string `json:"account_id"  validate:"required"`
}

// Validate ...
func (r *GetBPaasApplicationReq) Validate() error {
	return validator.Validate.Struct(r)
}
