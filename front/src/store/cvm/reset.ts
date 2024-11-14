import { defineStore } from 'pinia';
import { ref } from 'vue';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import type { IListResData } from '@/typings';
import http from '@/http';

export interface ICvmListRestStatus {
  id: string;
  account_id: string;
  bk_host_id: number;
  bk_host_name: string;
  bk_asset_id: string;
  private_ipv4_addresses: string[];
  private_ipv6_addresses: string[];
  public_ipv4_addresses: string[];
  public_ipv6_addresses: string[];
  operator: string;
  bak_operator: string;
  device_type: string;
  region: string;
  vendor: string;
  zone: string;
  bk_os_name: string;
  topo_module: string;
  bk_svr_source_type_id: string; // '1','2','3'是物理机, '4','5'是虚拟机
  status: string;
  srv_status: string;
  reset_status: number;
  // view-properties
  private_ip_address?: string;
  public_ip_address?: string;
  image_name_old?: string;
  cloud_image_id?: string;
  image_name?: string;
  image_type?: string;
}

interface ICvmBatchResetAsyncHost {
  id: string;
  bk_asset_id: string;
  device_type: string;
  image_name_old: string;
  cloud_image_id: string;
  image_name: string;
  image_type: string;
}

export const useCvmResetStore = defineStore('cvm-reset', () => {
  const { getBusinessApiPath } = useWhereAmI();

  // 查询虚拟机重装状态列表：docs/api-docs/web-server/docs/biz/list_cvm_reset_status.md
  const isCvmListResetStatusLoading = ref(false);
  const getCvmListResetStatus = async (data: { ids: string[] }) => {
    isCvmListResetStatusLoading.value = true;
    try {
      const res: IListResData<ICvmListRestStatus[]> = await http.post(
        `/api/v1/cloud/${getBusinessApiPath()}cvms/list/reset/status`,
        data,
      );
      return res?.data?.details || [];
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isCvmListResetStatusLoading.value = false;
    }
  };

  // 批量重装虚拟机：docs/api-docs/web-server/docs/biz/batch_reset_cvm.md
  const isCvmBatchResetAsyncLoading = ref(false);
  const cvmBatchResetAsync = async (params: { hosts: ICvmBatchResetAsyncHost[]; pwd: string; pwd_confirm: string }) => {
    isCvmBatchResetAsyncLoading.value = true;
    try {
      const res = await http.post(`/api/v1/cloud/${getBusinessApiPath()}cvms/batch/reset_async`, params);
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isCvmBatchResetAsyncLoading.value = false;
    }
  };

  return {
    isCvmListResetStatusLoading,
    getCvmListResetStatus,
    isCvmBatchResetAsyncLoading,
    cvmBatchResetAsync,
  };
});
