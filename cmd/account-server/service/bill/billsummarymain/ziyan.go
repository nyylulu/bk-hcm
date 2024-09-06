package billsummarymain

import (
	"hcm/pkg/api/core"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/finops"
	"hcm/pkg/tools/slice"
)

func (s *service) listProductName(kt *kit.Kit, productIds []int64) (map[int64]string, error) {
	productIds = slice.Unique(productIds)
	if len(productIds) == 0 {
		return nil, nil
	}
	productNameMap := make(map[int64]string)
	for _, ids := range slice.Split(productIds, int(filter.DefaultMaxInLimit)) {
		param := &finops.ListOpProductParam{
			OpProductIds: ids,
			Page:         *core.NewDefaultBasePage(),
		}
		productResult, err := s.finops.ListOpProduct(kt, param)
		if err != nil {
			logs.Errorf("list op product failed, productIDs: %v, err: %v, rid: %s", ids, err, kt.Rid)
			return nil, err
		}
		for _, product := range productResult.Items {
			productNameMap[product.OpProductId] = product.OpProductName
		}
	}

	return productNameMap, nil
}
