package bill

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/validator"
)

// ProductSummaryListReq list request for product summary
type ProductSummaryListReq struct {
	BillYear     int            `json:"bill_year" validate:"required"`
	BillMonth    int            `json:"bill_month" validate:"required"`
	OpProductIDs []int64        `json:"op_product_ids" validate:"required"`
	Page         *core.BasePage `json:"page" validate:"omitempty"`
}

// Validate ...
func (req *ProductSummaryListReq) Validate() error {
	return validator.Validate.Struct(req)
}
