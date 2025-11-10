<script setup lang="ts">
import { computed, inject, Ref, ref } from 'vue';
import useCvmChargeType from '@/views/ziyanScr/hooks/use-cvm-charge-type';
import ChargeMonthsSelector from '@/views/ziyanScr/cvm-produce/component/create-order/children/charge-months-selector.vue';
import { type ICvmDevicetypeItem } from '@/store/cvm/device';
import { RequirementType } from '@/store/config/requirement';
import { AvailableDeviceTypeMap } from './use-device-type-plan';

const chargeType = defineModel<string>('chargeType');

const chargeMonths = defineModel<number>('chargeMonths');

const props = defineProps<{
  availableDeviceTypeMap: AvailableDeviceTypeMap;
  selectedDeviceTypeList: ICvmDevicetypeItem[];
  isDefaultFourYears: boolean;
  isGpuDeviceType: boolean;
  isChargeTypeLoading: boolean;
}>();

const emit = defineEmits<{
  change: [type: string];
}>();

const { cvmChargeTypes, cvmChargeTypeNames } = useCvmChargeType();

const requireType = inject<RequirementType>('requireType');
const isRollingServer = inject<Ref<boolean>>('isRollingServer');
const isRollingServerOrGreenChannel = inject<Ref<boolean>>('isRollingServerOrGreenChannel');

const chargeMonthsDisabledState = computed(() => {
  if (props.isGpuDeviceType) {
    // GPU机型属于专用机型的特殊情况，只能选择6年
    return { disabled: true, content: 'GPU机型只能选择6年套餐' };
  }
  if (isRollingServer.value || props.isDefaultFourYears) {
    return {
      disabled: true,
      content: isRollingServer.value ? '继承原有套餐包年包月时长，此处的购买时长为剩余时长' : '专用机型只能选择4年套餐',
    };
  }

  return { disabled: false, content: '' };
});

const popConfirmRef = ref(null);

let confirmPromise = Promise.withResolvers();

const popConfirmProps = {
  trigger: 'manual',
  title: '确认切换计费模式？',
  content: '切换计费模式将导致所选内容被清空，请谨慎操作！',
  placement: 'top',
};

const handleTypeBeforeChange = () => {
  popConfirmRef.value?.$refs?.popoverRef?.show();
  return confirmPromise.promise
    .then(() => true)
    .catch(() => false)
    .finally(() => {
      confirmPromise = Promise.withResolvers();
    });
};

const handleTypeChange = (type: string) => {
  if (type === cvmChargeTypes.POSTPAID_BY_HOUR) {
    chargeMonths.value = undefined;
  } else {
    chargeMonths.value = 36;
  }
  emit('change', type);
};
</script>
<template>
  <div class="charge-type">
    <div class="item">
      <div class="form-label required title">计费模式</div>
      <bk-pop-confirm
        v-bind="popConfirmProps"
        @confirm="() => confirmPromise.resolve(1)"
        @cancel="() => confirmPromise.reject()"
        ref="popConfirmRef"
      >
        <!-- 滚服、小额绿通与预测无关 -->
        <bk-radio-group
          v-if="isRollingServerOrGreenChannel"
          type="card"
          class="radio-group"
          v-model="chargeType"
          :with-validate="false"
          :disabled="isRollingServer"
          v-bk-tooltips="{
            content: '继承原有套餐，计费模式不可选',
            disabled: !isRollingServer,
          }"
          :before-change="handleTypeBeforeChange"
          @change="handleTypeChange"
        >
          <bk-radio-button :label="cvmChargeTypes.PREPAID">
            {{ cvmChargeTypeNames[cvmChargeTypes.PREPAID] }}
          </bk-radio-button>
          <bk-radio-button :label="cvmChargeTypes.POSTPAID_BY_HOUR">
            {{ cvmChargeTypeNames[cvmChargeTypes.POSTPAID_BY_HOUR] }}
          </bk-radio-button>
        </bk-radio-group>

        <!-- 其它需求类型，需要看预测 -->
        <bk-radio-group
          v-else
          type="card"
          class="radio-group"
          v-model="chargeType"
          :with-validate="false"
          :disabled="isChargeTypeLoading"
          :before-change="handleTypeBeforeChange"
          @change="handleTypeChange"
        >
          <bk-radio-button
            :label="cvmChargeTypes.PREPAID"
            :disabled="availableDeviceTypeMap?.get(cvmChargeTypes.PREPAID)?.size === 0"
            v-bk-tooltips="{
              content: '当前地域无有效的预测需求，请提预测单后再按量申请',
              disabled: isChargeTypeLoading || availableDeviceTypeMap?.get(cvmChargeTypes.PREPAID)?.size > 0,
            }"
          >
            {{ cvmChargeTypeNames[cvmChargeTypes.PREPAID] }}
          </bk-radio-button>
          <bk-radio-button
            :label="cvmChargeTypes.POSTPAID_BY_HOUR"
            :disabled="availableDeviceTypeMap?.get(cvmChargeTypes.POSTPAID_BY_HOUR)?.size === 0"
            v-bk-tooltips="{
              content: '当前地域无有效的预测需求，请提预测单后再按量申请',
              disabled: isChargeTypeLoading || availableDeviceTypeMap?.get(cvmChargeTypes.POSTPAID_BY_HOUR)?.size > 0,
            }"
          >
            {{ cvmChargeTypeNames[cvmChargeTypes.POSTPAID_BY_HOUR] }}
          </bk-radio-button>
        </bk-radio-group>
      </bk-pop-confirm>
    </div>
    <div class="item" v-if="chargeType === cvmChargeTypes.PREPAID">
      <div class="form-label required title">购买时长</div>
      <charge-months-selector
        v-model="chargeMonths"
        :require-type="requireType"
        :is-gpu-device-type="isGpuDeviceType"
        :disabled="chargeMonthsDisabledState?.disabled"
        :with-validate="false"
        v-bk-tooltips="{
          content: chargeMonthsDisabledState?.content,
          disabled: !chargeMonthsDisabledState?.disabled,
        }"
      />
    </div>
  </div>
</template>

<style scoped lang="scss">
.charge-type {
  display: flex;
  flex-direction: column;
  gap: 16px;

  .radio-group {
    background: #fff;
  }

  .item {
    .title {
      margin-bottom: 8px;
    }
  }
}
</style>
