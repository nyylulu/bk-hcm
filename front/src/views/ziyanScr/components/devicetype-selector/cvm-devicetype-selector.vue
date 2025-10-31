<template>
  <devicetype-selector
    :class="selectorClass"
    ref="device-type-selector"
    v-model="model"
    resource-type="cvm"
    :params="params"
    :disabled="disabled"
    :loading="isLoading"
    :placeholder="placeholder"
    :sort="deviceTypeCompareFn"
    :option-disabled="deviceTypeOptionDisabledCallback"
    :option-disabled-tips-content="deviceTypeOptionDisabledTipsCallback"
    @change="handleChange"
    v-bind="attrs"
  >
    <template #option="option">
      <span>{{ option.device_type }}</span>
      <bk-tag class="ml12" :theme="option.device_type_class === 'SpecialType' ? 'danger' : 'success'" size="small">
        {{ option.device_type_class === 'SpecialType' ? '专用机型' : '通用机型' }}
      </bk-tag>
      <bk-tag v-if="option.device_group" class="ml12" size="small">{{ option.device_group }}</bk-tag>
    </template>
  </devicetype-selector>
  <cvm-devicetype-tip
    v-if="showTip && selectedCvmDeviceType"
    :class="tipClass"
    :info="selectedCvmDeviceType"
    :is-default-four-years="isDefaultFourYears"
    :is-gpu-device-type="isGpuDeviceType"
  >
    <template #default>
      <p v-if="isGreenChannel">
        注意：
        <span style="color: red">
          交付机型可能和所选不同，公司交付策略为同机型有资源优先交付，模糊机型范围为S4m、S5、S5t、S6、S6t、SA2、SA3、SA5t、SA5、SA6、SA9、S9
        </span>
      </p>
    </template>
  </cvm-devicetype-tip>
</template>

<script setup lang="ts">
import { computed, ref, useAttrs, useTemplateRef, watch } from 'vue';
import useCvmChargeType from '@/views/ziyanScr/hooks/use-cvm-charge-type';
import type { CvmDeviceType, DeviceType, SelectionType } from './types';
import type { RollingServerHost } from '../../rolling-server/inherit-package-form-item/index.vue';

import DevicetypeSelector from '@/views/ziyanScr/components/devicetype-selector/index.vue';
import CvmDevicetypeTip from '@/views/ziyanScr/components/devicetype-selector/cvm-tip.vue';
import { RequirementType } from '@/store/config/requirement';

interface IProps {
  region: string;
  zone: string;
  requireType: number;
  chargeType: string;
  computedAvailableDeviceTypeSet: Set<string>;
  rollingServerHost: RollingServerHost;
  disabled: boolean;
  isLoading: boolean;
  placeholder: string;
  selectorClass?: string;
  showTip?: boolean;
  tipClass?: string;
}

const model = defineModel<string>();
const props = defineProps<IProps>();
const emit = defineEmits<{
  change: [
    {
      deviceType: CvmDeviceType;
      chargeMonths: number;
      chargeMonthsDisabledState: { disabled: boolean; content: string };
    },
  ];
}>();
const attrs = useAttrs();

const selectorRef = useTemplateRef<typeof DevicetypeSelector>('device-type-selector');
const { cvmChargeTypes } = useCvmChargeType();

const params = computed(() => {
  const { region, zone, requireType } = props;
  return {
    region,
    zone: zone !== 'cvm_separate_campus' ? zone : undefined,
    require_type: requireType,
  };
});
const isRollingServer = computed(() => props.requireType === RequirementType.RollServer);
const isGreenChannel = computed(() => props.requireType === RequirementType.GreenChannel);
const isSpecialRequirement = computed(() =>
  [RequirementType.GreenChannel, RequirementType.RollServer].includes(props.requireType),
);

// 机型排序
const deviceTypeCompareFn = (a: DeviceType, b: DeviceType) => {
  // 非滚服、非小额绿通，走预测
  if (!isSpecialRequirement.value) {
    const set = props.computedAvailableDeviceTypeSet;
    return Number(set.has(b.device_type)) - Number(set.has(a.device_type));
  }
  // 滚服、小额绿通
  const {
    device_type_class: aDeviceTypeClass,
    device_group: aDeviceGroup,
    cpu_amount: aCpuAmount,
  } = a as CvmDeviceType;

  const {
    device_type_class: bDeviceTypeClass,
    device_group: bDeviceGroup,
    cpu_amount: bCpuAmount,
  } = b as CvmDeviceType;

  if (aDeviceTypeClass === 'CommonType' && bDeviceTypeClass === 'SpecialType') return -1;
  if (aDeviceTypeClass === 'SpecialType' && bDeviceTypeClass === 'CommonType') return 1;

  // 对小额绿通有特殊限制
  if (isGreenChannel.value) {
    const aDeviceValid = aDeviceGroup === '标准型' && aCpuAmount <= 16;
    const bDeviceValid = bDeviceGroup === '标准型' && bCpuAmount <= 16;
    return Number(bDeviceValid) - Number(aDeviceValid);
  }
  return 0;
};

// 机型选项禁用
const deviceTypeOptionDisabledCallback = (option: DeviceType) => {
  // 非滚服、非小额绿通
  if (!isSpecialRequirement.value) {
    return !props.computedAvailableDeviceTypeSet.has(option.device_type);
  }
  // 滚服、小额绿通
  const { device_type_class, device_group, cpu_amount } = option as CvmDeviceType;

  return (
    device_type_class === 'SpecialType' ||
    (isRollingServer.value && device_group !== props.rollingServerHost?.device_group) ||
    (isGreenChannel.value && !(device_group === '标准型' && cpu_amount <= 16))
  );
};

// 机型选项禁用tip
const deviceTypeOptionDisabledTipsCallback = (option: DeviceType) => {
  // 非滚服、非小额绿通
  if (!isSpecialRequirement.value) return '当前机型不在有效预测范围内';
  // 滚服、小额绿通
  const { device_type_class, device_group, cpu_amount } = option as CvmDeviceType;

  if (device_type_class === 'SpecialType') return '专用机型不允许选择';
  if (isRollingServer.value && device_group !== props.rollingServerHost?.device_group) return '机型族不匹配';
  if (isGreenChannel.value && !(device_group === '标准型' && cpu_amount <= 16)) return '非S类小核心不允许选择';
};

const selectedCvmDeviceType = ref<CvmDeviceType>(null);

// 常规项目-包年包月，专用机型默认包4年
const isDefaultFourYears = computed(
  () =>
    props.requireType === 1 &&
    props.chargeType === cvmChargeTypes.PREPAID &&
    selectedCvmDeviceType.value?.device_type_class === 'SpecialType',
);
// GPU机型默认包6年
const isGpuDeviceType = computed(
  () =>
    selectedCvmDeviceType.value?.device_type_class === 'SpecialType' &&
    selectedCvmDeviceType.value?.device_group.includes('GPU'),
);

const handleChange = (result: SelectionType) => {
  const deviceType = result as CvmDeviceType;
  selectedCvmDeviceType.value = deviceType;

  const { chargeMonths, chargeMonthsDisabledState } = calculateChargeMonthsState();

  emit('change', { deviceType, chargeMonths, chargeMonthsDisabledState });
};

// cvm机型的变更会影响到购买时长的禁用状态，由于依赖的状态都属于cvm-devicetype-selector组件，因此可对外暴露一个计算方法
const calculateChargeMonthsState = () => {
  const getTooltipOption = () => {
    if (isGpuDeviceType.value) {
      // GPU机型属于专用机型的特殊情况，只能选择6年
      return { disabled: true, content: 'GPU机型只能选择6年套餐' };
    }
    if (isRollingServer.value || isDefaultFourYears.value) {
      return {
        disabled: true,
        content: isRollingServer.value
          ? '继承原有套餐包年包月时长，此处的购买时长为剩余时长'
          : '专用机型只能选择4年套餐',
      };
    }
    return { disabled: false, content: '' };
  };

  // 计算购买时长
  let chargeMonths = 36;
  if (isDefaultFourYears.value) chargeMonths = 48;
  if (isGpuDeviceType.value) chargeMonths = 72;

  // 计算禁用状态
  const chargeMonthsDisabledState = getTooltipOption();

  return { chargeMonths, chargeMonthsDisabledState };
};

watch(model, (val) => {
  if (!val) {
    selectedCvmDeviceType.value = null;
  }
});

defineExpose({
  selectorRef,
  isGpuDeviceType,
  calculateChargeMonthsState,
});
</script>
