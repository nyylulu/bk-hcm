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
  IVerifyResourceDemandData,
  IVerifyResourceDemandParams,
} from '@/typings/plan';
import { IPlanTicketDemand } from '@/typings/resourcePlan';
import { defineStore } from 'pinia';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineStore('planStore', () => {
  const { getBusinessApiPath } = useWhereAmI();
  /**
   * 查询业务下资源预测需求列表
   */
  const list_biz_resource_plan_demand = async (
    ids: number[], // 预测需求IDS
  ): Promise<{
    [key: string]: any;
    data: {
      details: Partial<IDemandListDetail>[];
      [key: string]: any;
    };
  }> => {
    http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/${getBusinessApiPath()}plans/resources/demands/list`, {
      crp_demand_ids: ids,
      page: {
        count: false,
        start: 0,
        limit: 500,
      },
    });
    return {
      data: {
        overview: {
          total_cpu_core: 1024,
          total_applied_core: 1024,
          in_plan_cpu_core: 512,
          in_plan_applied_cpu_core: 512,
          out_plan_cpu_core: 512,
          out_plan_applied_cpu_core: 512,
          expiring_cpu_core: 224,
        },
        details: [
          {
            crp_demand_id: 11111,
            bk_biz_id: 111,
            bk_biz_name: '业务',
            op_product_id: 222,
            op_product_name: '运营产品',
            status: 'locked',
            status_name: '变更中',
            demand_class: 'CVM',
            available_year_month: '2024-01',
            expect_time: '2024-01-01',
            device_class: '高IO型I6t',
            device_type: 'I6t.33XMEDIUM198',
            total_os: 56,
            applied_os: 44,
            remained_os: 12,
            total_cpu_core: 560,
            applied_cpu_core: 440,
            remained_cpu_core: 120,
            total_memory: 560,
            applied_memory: 440,
            remained_memory: 120,
            total_disk_size: 560,
            applied_disk_size: 440,
            remained_disk_size: 120,
            region_id: 'ap-shanghai',
            region_name: '上海',
            zone_id: 'ap-shanghai-2',
            zone_name: '上海二区',
            plan_type: '预测内',
            obs_project: '常规项目',
            generation_type: '采购',
            device_family: '高IO型',
            disk_type: 'CLOUD_PREMIUM',
            disk_type_name: '高性能云硬盘',
            disk_io: 15,
          },
          {
            crp_demand_id: 222,
            bk_biz_id: 111,
            bk_biz_name: '业务',
            op_product_id: 222,
            op_product_name: '运营产品',
            status: 'locked',
            status_name: '变更中',
            demand_class: 'CVM',
            available_year_month: '2024-01',
            expect_time: '2024-01-01',
            device_class: '高IO型I6t',
            device_type: 'I6t.33XMEDIUM198',
            total_os: 56,
            applied_os: 44,
            remained_os: 12,
            total_cpu_core: 560,
            applied_cpu_core: 440,
            remained_cpu_core: 120,
            total_memory: 560,
            applied_memory: 440,
            remained_memory: 120,
            total_disk_size: 560,
            applied_disk_size: 440,
            remained_disk_size: 120,
            region_id: 'ap-shanghai',
            region_name: '上海',
            zone_id: 'ap-shanghai-2',
            zone_name: '上海二区',
            plan_type: '预测内',
            obs_project: '常规项目',
            generation_type: '采购',
            device_family: '高IO型',
            disk_type: 'CLOUD_PREMIUM',
            disk_type_name: '高性能云硬盘',
            disk_io: 15,
          },
        ],
      },
    };
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
    http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/findmany/config/cvm/charge_type/device_type`, data);
    return {
      data: {
        count: 2,
        info: [
          {
            charge_type: 'PREPAID',
            available: false,
            device_types: [
              {
                device_type: 'S3.6XLARGE64',
                available: true,
              },
              {
                device_type: 'S3.LARGE8',
                available: true,
              },
            ],
          },
          {
            charge_type: 'POSTPAID_BY_HOUR',
            available: true,
            device_types: [
              {
                device_type: 'S5.SMALL2',
                available: false,
              },
              {
                device_type: 'S5.LARGE16',
                available: false,
              },
            ],
          },
        ],
      },
    };
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
    http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plans/resources/demands/verify`, data);
    return {
      data: {
        verifications: [
          {
            verify_result: 'FAILED',
            reason: '',
          },
        ],
      },
    };
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
    http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/${getBusinessApiPath()}plans/resources/demands/adjust`, data);
    return {
      data: {
        id: '00000001',
      },
    };
  };

  /**
   * 查询期望交付时间对应的需求可用周范围及可用年月范围
   */
  const get_demand_available_time = async (
    expect_time: string,
  ): Promise<{
    data: IExceptTimeRange;
  }> => {
    http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plans/demands/available_times/get`, {
      expect_time,
    });
    return {
      data: {
        year_month_week: {
          year: 2024,
          month: 9,
          week_of_month: 5,
        },
        date_range_in_week: {
          start: '2024-09-30',
          end: '2024-10-06',
        },
        date_range_in_month: {
          start: '2024-10-28',
          end: '2024-11-03',
        },
      },
    };
  };

  /**
   * IDemandListDetail 转换为 IAdjust
   */
  function convertToAdjust(
    originalDetail: IDemandListDetail,
    updatedDetail: IDemandListDetail,
    adjustType: string,
    demandSource: string,
    expectTime?: string,
    delayReason?: string,
  ): IAdjust {
    const mapDetailToAdjustInfo = (detail: IDemandListDetail): AdjustInfo => {
      const demandResTypes: string[] = [];
      if (detail.total_cpu_core > 0 || detail.total_memory > 0) {
        demandResTypes.push('CVM');
      }
      if (detail.total_disk_size > 0) {
        demandResTypes.push('CBS');
      }

      return {
        obs_project: detail.obs_project,
        expect_time: detail.expect_time,
        region_id: detail.region_id,
        zone_id: detail.zone_id,
        demand_res_types: demandResTypes,
        cvm:
          detail.demand_class === 'CVM'
            ? {
                res_mode: detail.plan_type, // Assuming plan_type maps to res_mode
                device_type: detail.device_type,
                os: detail.total_os,
                cpu_core: detail.total_cpu_core,
                memory: detail.total_memory,
              }
            : undefined,
        cbs:
          detail.demand_class === 'CBS'
            ? {
                disk_type: detail.disk_type,
                disk_io: detail.disk_io,
                disk_size: detail.total_disk_size,
              }
            : undefined,
      };
    };

    return {
      crp_demand_id: originalDetail.crp_demand_id,
      adjust_type: adjustType,
      demand_source: demandSource,
      original_info: mapDetailToAdjustInfo(originalDetail),
      updated_info: mapDetailToAdjustInfo(updatedDetail),
      expect_time: adjustType === 'delay' ? expectTime : undefined,
      delay_reason: adjustType === 'delay' ? delayReason : undefined,
    };
  }

  /**
   * IDemandListDetail 转换为 IPlanTicketDemand
   */
  function convertToPlanTicketDemand(detail: IDemandListDetail): IPlanTicketDemand {
    const demand_res_types: string[] = [];
    if (detail.total_os > 0) demand_res_types.push('cvm');
    if (detail.total_disk_size > 0) demand_res_types.push('cbs');

    const cvm =
      detail.total_os > 0
        ? {
            res_mode: detail.plan_type,
            device_class: detail.device_class,
            device_type: detail.device_type,
            os: detail.total_os,
            cpu_core: detail.total_cpu_core,
            memory: detail.total_memory,
          }
        : undefined;

    const cbs =
      detail.total_disk_size > 0
        ? {
            disk_type: detail.disk_type,
            disk_type_name: detail.disk_type_name,
            disk_io: detail.disk_io,
            disk_size: detail.total_disk_size,
            disk_num: 1, // Assuming 1 for simplicity, adjust as needed
            disk_per_size: detail.total_disk_size, // Assuming total size for simplicity, adjust as needed
          }
        : undefined;

    return {
      obs_project: detail.obs_project,
      expect_time: detail.expect_time,
      region_id: detail.region_id,
      region_name: detail.region_name,
      zone_id: detail.zone_id,
      zone_name: detail.zone_name,
      demand_source: detail.demand_class,
      adjustType: detail.adjustType,
      crp_demand_id: detail.crp_demand_id,
      demand_res_types,
      cvm,
      cbs,
    };
  }

  /**
   * IPlanTicketDemand 转换为 IDemandListDetail
   */
  function convertToDemandListDetail(plan: IPlanTicketDemand, originDetail: IDemandListDetail): IDemandListDetail {
    const detail: IDemandListDetail = {
      crp_demand_id: plan.crp_demand_id, // 默认值，根据需要进行调整
      bk_biz_id: 0, // 默认值，根据需要进行调整
      bk_biz_name: '', // 默认值，根据需要进行调整
      op_product_id: 0, // 默认值，根据需要进行调整
      op_product_name: '', // 默认值，根据需要进行调整
      status: 'can_apply', // 默认值，根据需要进行调整
      status_name: '', // 默认值，根据需要进行调整
      demand_class: plan.demand_source,
      available_year_month: '', // 默认值，根据需要进行调整
      expect_time: plan.expect_time,
      device_class: plan.cvm?.device_class || '',
      device_type: plan.cvm?.device_type || '',
      total_os: plan.cvm?.os || 0,
      applied_os: 0, // 默认值，根据需要进行调整
      remained_os: plan.cvm?.os || 0, // 假设所有OS初始都剩余
      total_cpu_core: plan.cvm?.cpu_core || 0,
      applied_cpu_core: 0, // 默认值，根据需要进行调整
      remained_cpu_core: plan.cvm?.cpu_core || 0, // 假设所有CPU核数初始都剩余
      total_memory: plan.cvm?.memory || 0,
      applied_memory: 0, // 默认值，根据需要进行调整
      remained_memory: plan.cvm?.memory || 0, // 假设所有内存初始都剩余
      total_disk_size: plan.cbs?.disk_size || 0,
      applied_disk_size: 0, // 默认值，根据需要进行调整
      remained_disk_size: plan.cbs?.disk_size || 0, // 假设所有云盘大小初始都剩余
      region_id: plan.region_id,
      region_name: plan.region_name,
      zone_id: plan.zone_id,
      zone_name: plan.zone_name,
      plan_type: plan.cvm?.res_mode || '', // 假设cvm.res_mode为plan_type
      obs_project: plan.obs_project,
      generation_type: '', // 默认值，根据需要进行调整
      device_family: '', // 默认值，根据需要进行调整
      disk_type: plan.cbs?.disk_type || '',
      disk_type_name: plan.cbs?.disk_type_name || '',
      disk_io: plan.cbs?.disk_io || 0,
      adjustType: plan.adjustType,
    };
    if (JSON.stringify(detail) === JSON.stringify(originDetail)) detail.adjustType = AdjustType.none;
    return detail;
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
