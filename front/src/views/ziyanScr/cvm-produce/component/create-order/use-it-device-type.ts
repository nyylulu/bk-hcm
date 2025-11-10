import { computed, Ref, ref, watch } from 'vue';
import { ICloudInstanceConfigItem } from '@/typings/ziyanScr';
import { useZiyanScrStore } from '@/store';
import { useAccountSelectorStore } from '@/store/account-selector';
import { VendorEnum } from '@/common/constant';

export const useItDeviceType = (
  isBusiness: boolean,
  deviceType: Ref<string>,
  getParams: () => { region: string; zone: string | string[]; chargeType: string },
) => {
  const ziyanScrStore = useZiyanScrStore();
  const accountSelectorStore = useAccountSelectorStore();

  const currentCloudInstanceConfig = ref<ICloudInstanceConfigItem>();
  // 自研云账号只会有一个，这里直接通过store获取自研云账号
  const ziyanAccountId = computed(() => {
    const list = isBusiness ? accountSelectorStore.businessAccountList : accountSelectorStore.resourceAccountList;
    return list.find((item) => item.vendor === VendorEnum.ZIYAN)?.id ?? '';
  });
  const isItDeviceType = computed(() => /^(IT2|IT3(?!c)|I3|IT5|IT5c)/.test(deviceType.value));

  watch(deviceType, async (deviceType) => {
    if (!isItDeviceType.value) return;

    const { region, zone, chargeType } = getParams();

    // 兼容单可用区与多可用区，在cvm生产仍然使用的是单可用区
    let zoneFilters: Array<{ name: 'zone'; values: string[] }> = [];
    if (Array.isArray(zone)) {
      zoneFilters = zone?.[0] === 'all' ? [] : [{ name: 'zone', values: zone }];
    } else {
      zoneFilters = zone === 'cvm_separate_campus' ? [] : [{ name: 'zone', values: [zone] }];
    }

    const cloudInstanceConfigList = await ziyanScrStore.queryCloudInstanceConfig({
      account_id: ziyanAccountId.value,
      region,
      filters: [
        { name: 'instance-type', values: [deviceType] },
        { name: 'instance-charge-type', values: [chargeType] },
        ...zoneFilters,
      ],
    });

    // 只取第1条
    currentCloudInstanceConfig.value = cloudInstanceConfigList?.[0];
  });

  return {
    currentCloudInstanceConfig,
    ziyanAccountId,
    isItDeviceType,
  };
};
