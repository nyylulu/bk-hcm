/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package bkcc

import (
	"sync"

	clientv3 "go.etcd.io/etcd/client/v3"
	"hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/thirdparty/esb/cmdb"
)

// Syncer sync cc host operator
type Syncer struct {
	CliSet  *client.ClientSet
	EsbCli  esb.Client
	EtcdCli *clientv3.Client
	leaseOp *leaseOp
}

// NewSyncer create cc syncer
func NewSyncer(cliSet *client.ClientSet, esbCli esb.Client, etcdCli *clientv3.Client) (Syncer, error) {
	op := &leaseOp{cli: clientv3.NewLease(etcdCli), leaseMap: make(map[string]clientv3.LeaseID)}

	return Syncer{CliSet: cliSet, EsbCli: esbCli, EtcdCli: etcdCli, leaseOp: op}, nil
}

type leaseOp struct {
	sync.Mutex
	cli      clientv3.Lease
	leaseMap map[string]clientv3.LeaseID
}

func (l *leaseOp) getLeaseID(kt *kit.Kit, key string) (clientv3.LeaseID, error) {
	l.Lock()
	defer l.Unlock()

	leaseID, ok := l.leaseMap[key]
	var err error
	if ok {
		if _, err = l.cli.KeepAliveOnce(kt.Ctx, leaseID); err != nil {
			logs.Errorf("keep lease alive failed, err: %v, key: %s, leaseID: %v, rid: %s", err, key, leaseID, kt.Rid)
		}
	}

	if !ok || err != nil {
		var seconds int64 = 60 * 60
		leaseResp, err := l.cli.Grant(kt.Ctx, seconds)
		if err != nil {
			logs.Errorf("grant lease failed, err: %v, key: %s, rid: %s", err, key, kt.Rid)
			return 0, err
		}

		l.leaseMap[key] = leaseResp.ID
	}

	return l.leaseMap[key], nil
}

type ccHostWithBiz struct {
	cmdb.Host
	bizID int64
}

// getHostWithBizID 由于cc的主机模型没有业务id,所以这里需要会给主机信息补充业务id
func getHostWithBizID(bizID int64, hosts []cmdb.Host) []ccHostWithBiz {
	result := make([]ccHostWithBiz, 0)
	for _, host := range hosts {
		result = append(result, ccHostWithBiz{host, bizID})
	}

	return result
}

type diffHost struct {
	addHosts    []cloud.CvmBatchCreate[cvm.TCloudZiyanHostExtension]
	updateHosts []cloud.CvmBatchUpdate[cvm.TCloudZiyanHostExtension]
	deleteIDs   []string
}
