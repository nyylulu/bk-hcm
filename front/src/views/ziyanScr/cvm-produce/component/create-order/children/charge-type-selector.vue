<script setup lang="ts">
import { computed, useAttrs } from 'vue';
import { useI18n } from 'vue-i18n';
import useCvmChargeType from '@/views/ziyanScr/hooks/use-cvm-charge-type';
import { IOverflowTooltipProp } from 'bkui-vue/lib/table/props';

interface IProps {
  requireType: number;
  zone: string;
  availableDeviceTypeSet: Record<string, Set<any>>;
  disabled?: boolean;
  tooltipsOption?: IOverflowTooltipProp;
}

const model = defineModel<string>();
const props = withDefaults(defineProps<IProps>(), {
  disabled: false,
});
const attrs: any = useAttrs();
const { t } = useI18n();
const { cvmChargeTypes, cvmChargeTypeNames, cvmChargeTypeTips } = useCvmChargeType();

const isSpecialRequirement = computed(() => [6, 7].includes(props.requireType));
const chargeTypes = computed(() => {
  const baseTypes = [
    { value: cvmChargeTypes.PREPAID, name: cvmChargeTypeNames.PREPAID },
    { value: cvmChargeTypes.POSTPAID_BY_HOUR, name: cvmChargeTypeNames.POSTPAID_BY_HOUR },
  ];
  const tooltips = { content: t('当前地域无有效的预测需求，请提预测单后再按量申请'), disabled: true };

  if (isSpecialRequirement.value) {
    return baseTypes.map((type) => ({ ...type, disabled: false, tooltips }));
  }

  const { availableDeviceTypeSet, zone } = props;
  return baseTypes.map((type) => {
    const size = availableDeviceTypeSet[type.value]?.size || 0;
    return {
      ...type,
      disabled: size === 0,
      tooltips: { ...tooltips, disabled: !zone || size > 0 },
    };
  });
});
</script>

<template>
  <bk-radio-group v-model="model" :class="attrs.class" type="card" :disabled="disabled" v-bk-tooltips="tooltipsOption">
    <bk-radio-button
      v-for="chargeType in chargeTypes"
      :key="chargeType.value"
      :label="chargeType.value"
      :disabled="chargeType.disabled"
      v-bk-tooltips="chargeType.tooltips"
    >
      {{ chargeType.name }}
    </bk-radio-button>
  </bk-radio-group>
  <bk-alert theme="info" class="mt4">
    <template #title>
      {{ cvmChargeTypeTips[model] }}
      <bk-link href="https://crp.woa.com/crp-outside/yunti/news/20" theme="primary" target="_blank">
        <span class="font-small">{{ t('计费模式说明') }}</span>
      </bk-link>
    </template>
  </bk-alert>
</template>

<style lang="scss" scoped></style>
