/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package ziyan

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	apicore "hcm/pkg/api/core"
	datacli "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/json"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/util"
)

const (
	// TagKeyOpProduct 自研云 `运营产品` 标签名，对应值格式: `name_id`
	TagKeyOpProduct = "运营产品"
	// TagKeyBs1 自研云 `一级业务` 标签名，对应值格式: `name_id`
	TagKeyBs1 = "一级业务"
	// TagKeyBs2 自研云 `二级业务` 标签名，对应值格式: `name_id`
	TagKeyBs2 = "二级业务"
	// TagKeyOpDept 自研云 `虚拟部门` 标签名，对应值格式: `name_id`
	TagKeyOpDept = "运营部门"
	// TagKeyManager 自研云 `负责人` 标签名
	TagKeyManager = "负责人"
	// TagKeyBakManager 自研云 `备份负责人` 标签名
	TagKeyBakManager = "备份负责人"
)

// Bs2BizConfig 二级业务配置信息，对应 global_config 表中的配置数据
type Bs2BizConfig struct {
	OpProductID     int64  `json:"op_product_id"`
	OpProductName   string `json:"op_product_name"`
	Bs1NameID       int64  `json:"bs1_name_id"`
	Bs1Name         string `json:"bs1_name"`
	Bs2NameID       int64  `json:"bs2_name_id"`
	Bs2Name         string `json:"bs2_name"`
	VirtualDeptID   int64  `json:"virtual_dept_id"`
	VirtualDeptName string `json:"virtual_dept_name"`
	Manager         string `json:"manager"`
	BakManager      string `json:"bak_manager"`
	BkBizID         int64  `json:"bk_biz_id"`
}

// NotFoundID 未分配二级业务ID
const NotFoundID = -1

// ResourceMeta 资源归属元数据
type ResourceMeta struct {
	// 运营产品ID
	OpProductID int64 `json:"op_product_id,omitempty" validate:"required"`
	// 运营产品名称
	OpProductName string `json:"op_product_name,omitempty" validate:"required"`
	// 一级业务名
	Bs1Name string `json:"bs1_name,omitempty" validate:"required"`
	// 一级业务ID
	Bs1NameID int64 `json:"bs1_name_id,omitempty" validate:"required"`
	// 二级业务名
	Bs2Name string `json:"bs2_name,omitempty" validate:"required"`
	// 二级业务ID
	Bs2NameID int64 `json:"bs2_name_id,omitempty" validate:"required"`

	// 虚拟部门名称
	VirtualDeptName string `json:"virtual_dept_name,omitempty" validate:"required"`
	// 虚拟部门ID
	VirtualDeptID int64 `json:"virtual_dept_id,omitempty" validate:"required"`

	// 负责人
	Manager string `json:"manager,omitempty" validate:"required"`
	// 备份负责人
	BakManager string `json:"bak_manager,omitempty" validate:"omitempty"`
}

// Validate ...
func (i *ResourceMeta) Validate() error {
	return validator.Validate.Struct(i)
}

// GetSyncFilterTag 按二级业务标签过滤，带`tag:`前缀
func (i *ResourceMeta) GetSyncFilterTag() apicore.TagPair {
	return apicore.TagPair{Key: TagKeyBs2, Value: fmt.Sprintf("%s_%d", i.Bs2Name, i.Bs2NameID)}
}

// GetTagMap ...
func (i *ResourceMeta) GetTagMap() apicore.TagMap {

	tagMap := apicore.TagMap{
		TagKeyOpProduct: fmt.Sprintf("%s_%d", i.OpProductName, i.OpProductID),
		TagKeyBs1:       fmt.Sprintf("%s_%d", i.Bs1Name, i.Bs1NameID),
		TagKeyBs2:       fmt.Sprintf("%s_%d", i.Bs2Name, i.Bs2NameID),
		TagKeyOpDept:    fmt.Sprintf("%s_%d", i.VirtualDeptName, i.VirtualDeptID),
		TagKeyManager:   i.Manager,
	}
	if len(i.BakManager) > 0 {
		tagMap[TagKeyBakManager] = i.BakManager
	}
	return tagMap
}

// GetTagPairs ...
func (i *ResourceMeta) GetTagPairs() []apicore.TagPair {
	tagPairs := make([]apicore.TagPair, 0)

	if i.OpProductID > 0 {
		tagPairs = append(tagPairs, apicore.TagPair{Key: TagKeyOpProduct, Value: fmt.Sprintf("%s_%d",
			i.OpProductName, i.OpProductID)})
		tagPairs = append(tagPairs, apicore.TagPair{Key: TagKeyBs1, Value: fmt.Sprintf("%s_%d",
			i.Bs1Name, i.Bs1NameID)})
		tagPairs = append(tagPairs, apicore.TagPair{Key: TagKeyBs2, Value: fmt.Sprintf("%s_%d",
			i.Bs2Name, i.Bs2NameID)})
		tagPairs = append(tagPairs, apicore.TagPair{Key: TagKeyOpDept, Value: fmt.Sprintf("%s_%d",
			i.VirtualDeptName, i.VirtualDeptID)})
	}

	if len(i.Manager) > 0 {
		tagPairs = append(tagPairs, apicore.TagPair{Key: TagKeyManager, Value: i.Manager})
	}

	if len(i.BakManager) > 0 {
		tagPairs = append(tagPairs, apicore.TagPair{Key: TagKeyBakManager, Value: i.BakManager})
	}
	return tagPairs
}

// GetResourceMetaByBizForUser 为自研云资源生成业务标签
// 优先从 global_config 表中获取配置，如果没有则从 CMDB 查询
func GetResourceMetaByBizForUser(kt *kit.Kit, dataCli *datacli.Client, ccCli cmdb.Client, bkBizId int64,
	manager, bakManager string) (tags *ResourceMeta, err error) {

	if err = ensureInitialized(kt, dataCli); err != nil {
		// 读取初始化配置失败，继续使用 CMDB 查询
		logs.Warnf("failed to ensure bs2 to biz map initialized, will query from cc, bkBizId: %d, err: %v, "+
			"manager: %s, rid: %s", bkBizId, err, manager, kt.Rid)
	}

	// 1. 优先从内存缓存中查找配置
	if configVal, ok := bkBizIDToConfigMap.Load(bkBizId); ok {
		if config, ok := configVal.(*Bs2BizConfig); ok {
			// 使用配置中的 manager 和 bakManager，如果传入的参数为空
			finalManager := manager
			finalBakManager := bakManager
			if finalManager == "" && config.Manager != "" {
				finalManager = config.Manager
			}
			if finalBakManager == "" && config.BakManager != "" {
				finalBakManager = config.BakManager
			}

			meta := &ResourceMeta{
				OpProductID:     config.OpProductID,
				OpProductName:   config.OpProductName,
				Bs1Name:         config.Bs1Name,
				Bs1NameID:       config.Bs1NameID,
				Bs2Name:         config.Bs2Name,
				Bs2NameID:       config.Bs2NameID,
				VirtualDeptName: config.VirtualDeptName,
				VirtualDeptID:   config.VirtualDeptID,
				Manager:         finalManager,
				BakManager:      finalBakManager,
			}
			logs.Infof("found biz config from global_config cache for bk_biz_id: %d, meta: %+v, rid: %s",
				bkBizId, meta, kt.Rid)
			return meta, nil
		}
	}

	// 2. 如果缓存中没有，则从 CMDB 查询
	req := &cmdb.SearchBizCompanyCmdbInfoParams{BizIDs: []int64{bkBizId}}
	companyInfoList, err := ccCli.SearchBizCompanyCmdbInfo(kt, req)
	if err != nil {
		logs.Errorf("fail call cc to SearchBizCompanyCmdbInfo for cc biz id: %d, err: %v, rid: %s",
			bkBizId, err, kt.Rid)
		return nil, err
	}
	if companyInfoList == nil || len(*companyInfoList) < 1 {
		return nil, errors.New("no data returned from cc")
	}
	cmdbBizInfo := (*companyInfoList)[0]
	if cmdbBizInfo.BkBizID != bkBizId {
		logs.Errorf("company cmdb biz info from cc mismatch, want: %d, got: %d, rid: %s",
			bkBizId, cmdbBizInfo.BkBizID, kt.Rid)
		return nil, fmt.Errorf("company cmdb biz info from cc mismatch, want: %d, got: %d",
			bkBizId, cmdbBizInfo.BkBizID)
	}
	meta := &ResourceMeta{
		OpProductID:     cmdbBizInfo.BkProductID,
		OpProductName:   cmdbBizInfo.BkProductName,
		Bs1Name:         cmdbBizInfo.Bs1Name,
		Bs1NameID:       cmdbBizInfo.Bs1NameID,
		Bs2Name:         cmdbBizInfo.Bs2Name,
		Bs2NameID:       cmdbBizInfo.Bs2NameID,
		VirtualDeptName: cmdbBizInfo.VirtualDeptName,
		VirtualDeptID:   cmdbBizInfo.VirtualDeptID,
		Manager:         manager,
		// 备份负责人默认和当前用户一致
		BakManager: bakManager,
	}
	logs.Infof("biz config not found in cache, query from cmdb for bk_biz_id: %d, meta: %+v, rid: %s",
		bkBizId, meta, kt.Rid)

	return meta, nil
}

// GetResourceMetaByBiz 为自研云资源生成业务标签，默认当前用户为负责人和备份负责人
func GetResourceMetaByBiz(kt *kit.Kit, dataCli *datacli.Client, ccCli cmdb.Client, bkBizId int64) (
	tags *ResourceMeta, err error) {

	return GetResourceMetaByBizForUser(kt, dataCli, ccCli, bkBizId, kt.User, kt.User)
}

// GetResourceMetaByBizWithManager 为自研云资源生成业务标签，支持传入负责人和备份负责人
func GetResourceMetaByBizWithManager(kt *kit.Kit, dataCli *datacli.Client, ccCli cmdb.Client, bkBizId int64,
	manager, bakManager string) (tags *ResourceMeta, err error) {

	return GetResourceMetaByBizForUser(kt, dataCli, ccCli, bkBizId, manager, bakManager)
}

// 业务id 不会变化，只会增加，使用内存缓存提升性能
var bs2bkBizIDMap sync.Map

// bkBizID 到 Bs2BizConfig 的反向映射，用于快速查找业务配置信息
var bkBizIDToConfigMap sync.Map

// GetBkBizIdByBs2 根据二级业务id，查询对应的业务id
// 优先从 global_config 表读取映射关系，如果没有则到 cc 查询
// 如果没有找到会返回 constant.UnassignedBiz
func GetBkBizIdByBs2(kt *kit.Kit, dataCli *datacli.Client, ccCli cmdb.Client, bs2NameIDs []int64) (
	bizIds []int64, err error) {

	// 赋值业务id
	bizIds = make([]int64, len(bs2NameIDs))
	notFoundIds := make([]int64, 0, 100)

	// 1. 先从内存缓存查找
	for _, bs2id := range slice.Unique(bs2NameIDs) {
		if bs2id < 0 {
			// 跳过无效的id
			continue
		}
		if _, ok := bs2bkBizIDMap.Load(bs2id); !ok {
			notFoundIds = append(notFoundIds, bs2id)
		}
	}

	// 2. 如果内存中没有，尝试从 global_config 表加载
	if len(notFoundIds) > 0 {
		if err = ensureInitialized(kt, dataCli); err != nil {
			// 读取初始化配置失败，继续使用 CMDB 查询
			logs.Warnf("failed to ensure bs2 to biz map initialized, will query from cc, err: %v, notFoundIds: %v, "+
				"rid: %s", err, notFoundIds, kt.Rid)
		}

		// 重新检查哪些还没找到
		stillNotFound := make([]int64, 0, len(notFoundIds))
		for _, bs2id := range notFoundIds {
			if _, ok := bs2bkBizIDMap.Load(bs2id); !ok {
				stillNotFound = append(stillNotFound, bs2id)
			}
		}
		notFoundIds = stillNotFound
	}

	// 3. 如果 global_config 中也没有，则从 cc 查询
	if len(notFoundIds) > 0 {
		param := &cmdb.SearchBizParams{
			Page: cmdb.BasePage{},
			BizPropertyFilter: &cmdb.QueryFilter{
				Rule: cmdb.Combined(cmdb.ConditionAnd, cmdb.In("bs2_name_id", notFoundIds)),
			},
		}
		business, err := ccCli.SearchBusiness(kt, param)
		if err != nil {
			logs.Errorf("fail to search cmdb business, err: %v,bs2_name_id list: %v, rid: %s", err, notFoundIds, kt.Rid)
			return nil, err
		}
		for _, biz := range business.Info {
			bs2bkBizIDMap.Store(biz.BsName2ID, biz.BizID)
		}
	}

	// 4. 填充结果
	for i := range bs2NameIDs {
		bs2NameID := bs2NameIDs[i]
		// 对没有解析到标签的业务，fallback到未分配
		bizIds[i] = constant.UnassignedBiz
		if bkBizID, ok := bs2bkBizIDMap.Load(bs2NameID); ok {
			bizIds[i], err = util.GetInt64ByInterface(bkBizID)
			// should never happen
			if err != nil {
				logs.Errorf("fail to convert bkBizID to int64, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		}
	}
	logs.Infof("get bk_biz_id by bs2 success, bs2NameIDs: %v, bizIds: %v, rid: %s", bs2NameIDs, bizIds, kt.Rid)
	return bizIds, nil
}

// 加载控制，确保配置只从数据库加载一次
var (
	loadOnce      sync.Once
	loadErr       error
	isInitialized bool
)

// ensureInitialized 确保配置已初始化，如果未初始化则尝试懒加载
// 这是一个内部辅助函数，用于在需要时自动加载配置
func ensureInitialized(kt *kit.Kit, dataCli *datacli.Client) error {
	if isInitialized {
		return nil
	}

	// 如果还没有初始化，尝试加载
	if dataCli != nil {
		return initBs2BizMap(kt, dataCli)
	}

	// 如果没有 dataCli，返回错误
	return fmt.Errorf("bs2 biz map not initialized and no data client available")
}

// initBs2BizMap 初始化 bs2 到 bk_biz_id 的映射，在服务启动时调用
// 从 global_config 表加载所有配置到内存缓存
// 此函数可以被多次调用，但只会执行一次加载
func initBs2BizMap(kt *kit.Kit, dataCli *datacli.Client) error {
	loadOnce.Do(func() {
		loadErr = loadBs2BizMapFromDB(kt, dataCli)
		if loadErr == nil {
			isInitialized = true
			logs.Infof("init bs2 to biz mapping success, rid: %s", kt.Rid)
		} else {
			logs.Errorf("init bs2 to biz mapping failed, err: %v, rid: %s", loadErr, kt.Rid)
		}
	})
	return loadErr
}

// loadBs2BizMapFromDB 从 global_config 表加载 bs2 到 bk_biz_id 的映射
// config_value 格式为 JSON 数组，包含完整的业务信息
// 同时加载正向映射（bs2_name_id -> bk_biz_id）和反向映射（bk_biz_id -> Bs2BizConfig）
func loadBs2BizMapFromDB(kt *kit.Kit, dataCli *datacli.Client) error {
	if dataCli == nil {
		return fmt.Errorf("failed to load bs2bizmap from db, data client is nil")
	}

	// 查询固定的配置记录
	listReq := &apicore.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", enumor.GlobalConfigTypeBs2ToBkBizIDMap),
			tools.RuleEqual("config_key", enumor.GlobalConfigKeyBs2BizMapping),
		),
		Page: apicore.NewDefaultBasePage(),
	}

	list, err := dataCli.Global.GlobalConfig.List(kt, listReq)
	if err != nil {
		logs.Errorf("failed to list bs2 biz config from global_config, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if len(list.Details) == 0 {
		logs.Warnf("no bs2 biz config found in global_config, rid: %s", kt.Rid)
		return nil
	}

	// 解析 JSON 数组格式的配置
	detail := list.Details[0]
	var configs []Bs2BizConfig
	if err = json.UnmarshalFromString(string(detail.ConfigValue), &configs); err != nil {
		logs.Errorf("failed to unmarshal bs2 biz config, value: %s, err: %v, rid: %s",
			detail.ConfigValue, err, kt.Rid)
		return err
	}

	// 将查询结果加载到内存缓存
	loadedCount := 0
	for _, config := range configs {
		bkBizID := config.BkBizID
		if bkBizID == 0 {
			logs.Warnf("bs2_name_id %d has no bk_biz_id in config, skip, rid: %s", config.Bs2NameID, kt.Rid)
			continue
		}

		// 正向映射：bs2_name_id -> bk_biz_id
		bs2bkBizIDMap.Store(config.Bs2NameID, bkBizID)

		// 反向映射：bk_biz_id -> Bs2BizConfig 复制一份避免循环变量问题
		configCopy := config
		bkBizIDToConfigMap.Store(bkBizID, &configCopy)

		loadedCount++
	}

	logs.Infof("loaded bs2 to biz mappings from global_config, loadNum: %d, totalNum: %d, configs: %+v, rid: %s",
		loadedCount, len(configs), configs, kt.Rid)
	return nil
}

// parseNameID 解析 xxx_yyy_123 格式的字符串，返回 xxx_yyy 和 123
func parseNameID(key string, val string) (name string, id int64, err error) {
	parts := strings.Split(val, "_")
	if len(parts) < 2 {
		return "", -1, fmt.Errorf("tag %s value %s format mismatch xxx_123", key, val)
	}
	name = strings.Join(parts[:len(parts)-1], "_")
	id, err = strconv.ParseInt(parts[len(parts)-1], 10, 64)
	return name, id, err
}

// ParseResourceMetaIgnoreErr 解析标签中的资源元数据，尽力解析，忽略错误，不校验完整性
func ParseResourceMetaIgnoreErr(resTagMap apicore.TagMap) (meta *ResourceMeta) {
	meta = NewResourceMeta()
	// ignore err，不会返回错误
	_ = parseResourceMeta(meta, resTagMap, true)
	return meta
}

// ParseResourceMeta 解析标签中的资源元数据, 严格模式
func ParseResourceMeta(resTagMap apicore.TagMap) (meta *ResourceMeta, err error) {
	meta = NewResourceMeta()
	// ignore err，不会返回错误

	if err = parseResourceMeta(meta, resTagMap, false); err != nil {
		return nil, err
	}

	if err = meta.Validate(); err != nil {
		return nil, fmt.Errorf("fail to validate meta: %v", err)
	}
	return meta, nil
}

// parseResourceMeta 解析标签中的资源元数据, 不保证信息完整性和准确性，依赖给定的标签。
// 入参 ignoreParseErr 为true时，会略解析错误，接续解析其他标签。目前只在存在某个标签，但是其值中的id字段解析失败时，会发生解析失败的情况
// 入参 ignoreParseErr 为false时，解析id失败会直接返回错误，中断解析
// 该函数不校验标签完整性，使用 ResourceMeta.Validate 检查完整性
func parseResourceMeta(meta *ResourceMeta, resTagMap apicore.TagMap, ignoreParseErr bool) (err error) {
	for key, value := range resTagMap {
		switch key {
		case TagKeyBs1:
			// FORMAT: Bs1Name_Bs2NameID
			name, id, err := parseNameID(key, value)
			if err != nil {
				logs.Warnf("fail to parse tag %s=%s, err: %v", key, value, err)
				if ignoreParseErr {
					break
				}
				return fmt.Errorf("fail to parse tag %s=%s, err: %v", key, value, err)
			}
			meta.Bs1Name = name
			meta.Bs1NameID = id
		case TagKeyBs2:
			// FORMAT: Bs1Name_Bs2NameID
			name, id, err := parseNameID(key, value)
			if err != nil {
				logs.Warnf("fail to parse tag %s=%s, err: %v", key, value, err)
				if ignoreParseErr {
					break
				}
				return fmt.Errorf("fail to parse tag %s=%s, err: %v", key, value, err)
			}
			meta.Bs2Name = name
			meta.Bs2NameID = id
		case TagKeyOpProduct:
			// FORMAT: OpProductName_OpProductID
			name, id, err := parseNameID(key, value)
			if err != nil {
				logs.Warnf("fail to parse tag %s=%s, err: %v", key, value, err)
				if ignoreParseErr {
					break
				}
				return fmt.Errorf("fail to parse tag %s=%s, err: %v", key, value, err)
			}
			meta.OpProductName = name
			meta.OpProductID = id
		case TagKeyOpDept:
			// FORMAT: VirtualDeptName_VirtualDeptID
			name, id, err := parseNameID(key, value)
			if err != nil {
				logs.Warnf("fail to parse tag %s=%s, err: %v", key, value, err)
				if ignoreParseErr {
					break
				}
				return fmt.Errorf("fail to parse tag %s=%s, err: %v", key, value, err)
			}
			meta.VirtualDeptName = name
			meta.VirtualDeptID = id
		case TagKeyManager:
			meta.Manager = value
		case TagKeyBakManager:
			meta.BakManager = value
		default:
			// 忽略其他标签
		}
	}
	return nil
}

// NewResourceMeta 新建资源元数据, 名字字段默认为空，ID默认为-1
func NewResourceMeta() *ResourceMeta {
	meta := &ResourceMeta{
		Bs1NameID:     NotFoundID,
		Bs2NameID:     NotFoundID,
		OpProductID:   NotFoundID,
		VirtualDeptID: NotFoundID,
	}
	return meta
}
