import { computed, reactive, Ref, ref } from 'vue';
import usePlanStore from '@/store/usePlanStore';
import useCvmChargeType from '@/views/ziyanScr/hooks/use-cvm-charge-type';
import type { DeviceType } from '@/views/ziyanScr/components/devicetype-selector/types';

import DevicetypeSelector from '@/views/ziyanScr/components/devicetype-selector/index.vue';

export default (selectorRef: Ref<typeof DevicetypeSelector>, chargeType: Ref<string>) => {
  const planStore = usePlanStore();
  const { cvmChargeTypes } = useCvmChargeType();

  const isPlanedDeviceTypeLoading = ref(false);
  const availableDeviceTypeSet = reactive<{ [key in string]: Set<any> }>({ prepaid: new Set(), postpaid: new Set() }); // 预测内的机型
  const hasPlanedDeviceType = computed(
    () => availableDeviceTypeSet.postpaid.size || availableDeviceTypeSet.prepaid.size,
  ); // 有预测内的机型

  // 用于禁用预测外的机型
  const computedAvailableDeviceTypeSet = computed(() => {
    const targetSet =
      chargeType.value === cvmChargeTypes.PREPAID ? availableDeviceTypeSet.prepaid : availableDeviceTypeSet.postpaid;

    // 将预测内的机型列表排在前面
    selectorRef.value.handleSort(
      (a: DeviceType, b: DeviceType) => Number(targetSet.has(b.device_type)) - Number(targetSet.has(a.device_type)),
    );
    return targetSet;
  });

  // 查询业务下的预测余量
  const getPlanedDeviceType = async (bk_biz_id: number, require_type: number, region: string, zone: string) => {
    isPlanedDeviceTypeLoading.value = true;
    availableDeviceTypeSet.prepaid.clear();
    availableDeviceTypeSet.postpaid.clear();

    try {
      const { data } = await planStore.list_config_cvm_charge_type_device_type({
        bk_biz_id,
        require_type,
        region,
        zone,
      });
      const { info } = data;

      info.forEach(({ charge_type, device_types }) => {
        const targetSet =
          cvmChargeTypes.PREPAID === charge_type ? availableDeviceTypeSet.prepaid : availableDeviceTypeSet.postpaid;

        device_types.forEach(({ device_type, available }) => {
          if (available) targetSet.add(device_type);
        });
      });

      return info;
    } finally {
      isPlanedDeviceTypeLoading.value = false;
    }
  };

  return {
    isPlanedDeviceTypeLoading,
    availableDeviceTypeSet,
    computedAvailableDeviceTypeSet,
    hasPlanedDeviceType,
    getPlanedDeviceType,
  };
};
