import { useWhereAmI } from '@/hooks/useWhereAmI';
import http from '@/http';
import {
  AdjustInfo,
  AdjustType,
  IAdjust,
  IAdjustData,
  IAdjustParams,
  IDemandListDetail,
  IExceptTimeRange,
  IListConfigCvmChargeTypeDeviceTypeData,
  IListConfigCvmChargeTypeDeviceTypeParams,
  ITimeRange,
  IVerifyResourceDemandData,
  IVerifyResourceDemandParams,
} from '@/typings/plan';
import { IPlanTicketDemand } from '@/typings/resourcePlan';
import { isNil, mergeWith } from 'lodash-es';
import { defineStore } from 'pinia';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineStore('planStore', () => {
  const { getBusinessApiPath } = useWhereAmI();
  /**
   * 查询业务下资源预测需求列表
   */
  const list_biz_resource_plan_demand = async (
    ids: string[], // 预测需求IDS
    expect_time_range: ITimeRange,
  ): Promise<{
    [key: string]: any;
    data: {
      details: Partial<IDemandListDetail>[];
      [key: string]: any;
    };
  }> => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/${getBusinessApiPath()}plans/resources/demands/list`, {
      demand_ids: ids,
      expect_time_range,
      page: {
        count: false,
        start: 0,
        limit: 500,
      },
    });
  };

  /**
   * 查询计费模式及机型配置信息
   */
  const list_config_cvm_charge_type_device_type = async (
    data: IListConfigCvmChargeTypeDeviceTypeParams,
  ): Promise<{
    [key: string]: any;
    data: IListConfigCvmChargeTypeDeviceTypeData;
  }> => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/findmany/config/cvm/charge_type/device_type`, data);
  };

  /**
   * 资源预测需求校验
   */
  const verify_resource_demand = async (
    data: IVerifyResourceDemandParams,
  ): Promise<{
    [key: string]: any;
    data: IVerifyResourceDemandData;
  }> => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plans/resources/demands/verify`, data);
  };

  /**
   * 批量调整资源预测需求
   */
  const adjust_biz_resource_plan_demand = async (
    data: IAdjustParams,
  ): Promise<{
    [key: string]: any;
    data: IAdjustData;
  }> => {
    return http.post(
      `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/${getBusinessApiPath()}plans/resources/demands/adjust`,
      data,
    );
  };

  /**
   * 查询期望交付时间对应的需求可用周范围及可用年月范围
   */
  const get_demand_available_time = async (
    expect_time: string,
  ): Promise<{
    data: IExceptTimeRange;
  }> => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plans/demands/available_times/get`, {
      expect_time,
    });
  };

  /**
   * IDemandListDetail 转换为 IAdjust
   */
  function convertToAdjust(
    originalDetail: IDemandListDetail,
    updatedDetail: IDemandListDetail,
    delayOs?: string,
  ): IAdjust {
    const mapDetailToAdjustInfo = (detail: IDemandListDetail): AdjustInfo => {
      const demandResTypes: string[] = ['CBS'];
      if (detail.remained_cpu_core > 0 || detail.remained_memory > 0) {
        demandResTypes.push('CVM');
      }

      return {
        obs_project: detail.obs_project,
        expect_time: detail.expect_time,
        region_id: detail.region_id,
        zone_id: detail.zone_id,
        demand_res_types: demandResTypes,
        // demand_source、remark 在编辑时不可调整，接口需要填写，产品结论是先给默认值
        demand_source: '指标变化',
        remark: '',
        cvm: {
          res_mode: detail.res_mode,
          device_type: detail.device_type,
          os: detail.remained_os,
          cpu_core: detail.remained_cpu_core,
          memory: detail.remained_memory,
        },
        cbs: {
          disk_type: detail.disk_type,
          disk_io: detail.disk_io,
          disk_size: detail.remained_disk_size,
        },
      };
    };

    return {
      demand_id: originalDetail.demand_id,
      adjust_type: updatedDetail.adjustType,
      demand_source: updatedDetail.demand_source,
      original_info: mapDetailToAdjustInfo(originalDetail),
      updated_info: mapDetailToAdjustInfo(updatedDetail),
      expect_time: updatedDetail.adjustType === AdjustType.time ? updatedDetail.expect_time : undefined,
      delay_reason: updatedDetail.adjustType === AdjustType.time ? '' : undefined,
      delay_os: delayOs,
    };
  }

  /**
   * IDemandListDetail 转换为 IPlanTicketDemand
   */
  function convertToPlanTicketDemand(detail: IDemandListDetail): IPlanTicketDemand {
    const demand_res_types: string[] = [detail.demand_res_type];
    if (detail.demand_res_type === 'CVM') demand_res_types.push('CBS');

    const cvm =
      +detail.remained_os > 0
        ? {
            res_mode: detail.res_mode,
            device_class: detail.device_class,
            device_type: detail.device_type,
            os: detail.remained_os,
            cpu_core: detail.remained_cpu_core,
            memory: detail.remained_memory,
          }
        : undefined;

    const cbs =
      detail.remained_disk_size > 0
        ? {
            disk_type: detail.disk_type,
            disk_type_name: detail.disk_type_name,
            disk_io: detail.disk_io,
            disk_size: detail.remained_disk_size,
            disk_per_size:
              detail.remained_disk_size && detail.remained_os
                ? Math.floor(detail.remained_disk_size / +detail.remained_os)
                : 0,
          }
        : undefined;

    return {
      obs_project: detail.obs_project,
      expect_time: detail.expect_time,
      region_id: detail.region_id,
      region_name: detail.region_name,
      zone_id: detail.zone_id,
      zone_name: detail.zone_name,
      demand_source: detail.demand_source,
      demand_class: detail.demand_class,
      adjustType: detail.adjustType,
      demand_id: detail.demand_id,
      demand_res_types,
      cvm,
      cbs,
    };
  }

  /**
   * IPlanTicketDemand 转换为 IDemandListDetail
   */
  function convertToDemandListDetail(plan: IPlanTicketDemand, originDetail: IDemandListDetail): IDemandListDetail {
    const detail: Partial<IDemandListDetail> = {
      demand_id: plan.demand_id,
      demand_class: plan.demand_class,
      demand_source: plan.demand_source,
      expect_time: plan.expect_time,
      device_class: plan.cvm?.device_class,
      device_type: plan.cvm?.device_type,
      remained_os: plan.cvm?.os,
      remained_cpu_core: plan.cvm?.cpu_core,
      remained_memory: plan.cvm?.memory,
      remained_disk_size: plan.cbs?.disk_size,
      region_id: plan.region_id,
      region_name: plan.region_name,
      zone_id: plan.zone_id,
      zone_name: plan.zone_name,
      obs_project: plan.obs_project,
      disk_type: plan.cbs?.disk_type,
      disk_type_name: plan.cbs?.disk_type_name,
      disk_io: plan.cbs?.disk_io,
      adjustType: plan.adjustType,
    };
    if (JSON.stringify(detail) === JSON.stringify(originDetail)) detail.adjustType = AdjustType.none;
    const res = mergeWith({}, originDetail, detail, (v1, v2) => (isNil(v2) ? v1 : v2));
    return res;
  }

  return {
    list_biz_resource_plan_demand,
    list_config_cvm_charge_type_device_type,
    verify_resource_demand,
    adjust_biz_resource_plan_demand,
    get_demand_available_time,
    // 工具函数
    convertToAdjust,
    convertToPlanTicketDemand,
    convertToDemandListDetail,
  };
});
