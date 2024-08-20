package sync

import (
	"fmt"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
)

// TCloudZiyanSyncHostReq tcloud ziyan sync host request.
type TCloudZiyanSyncHostReq struct {
	AccountID  string  `json:"account_id" validate:"required"`
	BizID      int64   `json:"bk_biz_id" validate:"required"`
	DelHostIDs []int64 `json:"delete_host_ids"`
}

// Validate ...
func (req *TCloudZiyanSyncHostReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudZiyanSyncHostByCondReq tcloud ziyan sync host by cond request.
type TCloudZiyanSyncHostByCondReq struct {
	AccountID string  `json:"account_id" validate:"required"`
	BizID     int64   `json:"bk_biz_id" validate:"required"`
	HostIDs   []int64 `json:"bk_host_ids"`
}

// Validate ...
func (req *TCloudZiyanSyncHostByCondReq) Validate() error {
	if len(req.HostIDs) > constant.CloudResourceSyncMaxLimit {
		return fmt.Errorf("host ids should <= %d", constant.CloudResourceSyncMaxLimit)
	}

	if req.BizID == 0 {
		return fmt.Errorf("bk_biz_id is invalid")
	}

	if req.AccountID == "" {
		return fmt.Errorf("account_id is invalid")
	}

	return nil
}

// TCloudZiyanDelHostByCondReq tcloud ziyan delete host by condition request.
type TCloudZiyanDelHostByCondReq struct {
	AccountID string  `json:"account_id" validate:"required"`
	HostIDs   []int64 `json:"bk_host_ids"`
}

// Validate ...
func (req *TCloudZiyanDelHostByCondReq) Validate() error {
	if len(req.HostIDs) > constant.CloudResourceSyncMaxLimit {
		return fmt.Errorf("host ids should <= %d", constant.CloudResourceSyncMaxLimit)
	}

	if req.AccountID == "" {
		return fmt.Errorf("account_id is invalid")
	}

	return nil
}
