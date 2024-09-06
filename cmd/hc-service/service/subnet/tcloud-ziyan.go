package subnet

import (
	"fmt"

	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	proto "hcm/pkg/api/hc-service/subnet"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// TCloudZiyanListSubnetCountIP count tcloud ziyan subnets' available ips.
func (s subnet) TCloudZiyanListSubnetCountIP(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.ListCountIPReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Page: core.NewDefaultBasePage(),
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: req.Region},
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: req.AccountID},
				&filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: req.IDs},
			},
		},
	}
	listResult, err := s.cs.DataService().Global.Subnet.List(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		return nil, err
	}

	if len(listResult.Details) != len(req.IDs) {
		return nil, fmt.Errorf("list subnet return count not right, query id count: %d, but return %d",
			len(req.IDs), len(listResult.Details))
	}

	cloudIDs := make([]string, 0, len(listResult.Details))
	for _, one := range listResult.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	cli, err := s.ad.TCloudZiyan(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	listOpt := &adcore.TCloudListOption{
		Region:   req.Region,
		CloudIDs: cloudIDs,
		Page: &adcore.TCloudPage{
			Offset: 0,
			Limit:  adcore.TCloudQueryLimit,
		},
	}
	subnetRes, err := cli.ListSubnet(cts.Kit, listOpt)
	if err != nil {
		return nil, err
	}

	if len(subnetRes.Details) != len(cloudIDs) {
		return nil, fmt.Errorf("list tcloud subnet return count not right, query id count: %d, but return %d",
			len(cloudIDs), len(subnetRes.Details))
	}

	cloudIDMap := make(map[string]string)
	for _, one := range listResult.Details {
		cloudIDMap[one.CloudID] = one.ID
	}

	result := make(map[string]proto.AvailIPResult)
	for _, one := range subnetRes.Details {
		id, exist := cloudIDMap[one.CloudID]
		if !exist {
			return nil, fmt.Errorf("subnet: %s not found", one.CloudID)
		}

		result[id] = proto.AvailIPResult{
			AvailableIPCount: one.Extension.AvailableIPAddressCount,
			TotalIPCount:     one.Extension.TotalIpAddressCount,
			UsedIPCount:      one.Extension.UsedIpAddressCount,
		}
	}

	return result, nil
}
