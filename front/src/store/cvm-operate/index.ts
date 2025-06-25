import { defineStore } from 'pinia';
import { ref } from 'vue';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import type { IListResData } from '@/typings';
import http from '@/http';

export interface ICvmListOperateStatus {
  id: string;
  vendor: string;
  account_id: string;
  bk_host_id: number;
  bk_host_name: string;
  cloud_id: string;
  bk_asset_id: string;
  private_ipv4_addresses: string[];
  private_ipv6_addresses: string[];
  public_ipv4_addresses: string[];
  public_ipv6_addresses: string[];
  operator: string;
  bak_operator: string;
  device_type: string;
  region: string;
  zone: string;
  bk_os_name: string;
  topo_module: string;
  bk_svr_source_type_id: string; // '1','2','3'是物理机, '4','5'是虚拟机
  status: string;
  srv_status: string;
  operate_status: number;
  // view-properties
  private_ip_address?: string;
  public_ip_address?: string;
  image_name_old?: string;
  cloud_image_id?: string;
  image_name?: string;
  image_type?: string;
}

export type CvmOperateType = 'start' | 'stop' | 'reboot' | 'reset';

// 批量重装
interface ICvmBatchResetAsyncHostItem {
  id: string;
  bk_asset_id: string;
  device_type: string;
  image_name_old: string;
  cloud_image_id: string;
  image_name: string;
  image_type: string;
}

export const useCvmOperateStore = defineStore('cvm-operate', () => {
  const { getBusinessApiPath } = useWhereAmI();

  // 查询虚拟机可操作状态列表, 如开关机、重启、重装：/docs/api-docs/web-server/docs/biz/cvm/list_cvm_operate_status.md
  const isCvmListOperateStatusLoading = ref(false);
  const getCvmListOperateStatus = async (data: { ids: string[]; operate_type: CvmOperateType }) => {
    isCvmListOperateStatusLoading.value = true;
    try {
      const res: IListResData<ICvmListOperateStatus[]> = await http.post(
        `/api/v1/cloud/${getBusinessApiPath()}cvms/list/operate/status`,
        data,
      );
      return res?.data?.details || [];
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isCvmListOperateStatusLoading.value = false;
    }
  };

  // 批量重装虚拟机：/docs/api-docs/web-server/docs/biz/cvm/batch_reset_cvm.md
  const isCvmBatchResetAsyncLoading = ref(false);
  const cvmBatchResetAsync = async (params: {
    hosts: ICvmBatchResetAsyncHostItem[];
    pwd: string;
    pwd_confirm: string;
    session_id: string;
  }) => {
    isCvmBatchResetAsyncLoading.value = true;
    try {
      const res = await http.post(`/api/v1/cloud/${getBusinessApiPath()}cvms/batch/reset_async`, params, {
        globalError: false,
      });
      return res;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isCvmBatchResetAsyncLoading.value = false;
    }
  };

  return {
    isCvmListOperateStatusLoading,
    getCvmListOperateStatus,
    isCvmBatchResetAsyncLoading,
    cvmBatchResetAsync,
  };
});
