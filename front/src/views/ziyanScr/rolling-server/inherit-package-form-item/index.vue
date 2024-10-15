<script setup lang="ts">
import { ref, useTemplateRef } from 'vue';
import { Form } from 'bkui-vue';
import InputWithValidate from '@/components/input-with-validate/index.vue';

import { useI18n } from 'vue-i18n';
import { useWhereAmI, Senarios } from '@/hooks/useWhereAmI';
import { timeFormatter } from '@/common/util';
import { INSTANCE_CHARGE_MAP } from '@/common/constant';
import http from '@/http';
import useCvmChargeType from '../../hooks/use-cvm-charge-type';

const { FormItem } = Form;

export interface RollingServerHost {
  device_type: string;
  instance_charge_type: string;
  charge_months: number;
  billing_start_time: string;
  old_billing_expire_time: string;
  bk_cloud_inst_id: string;
}

defineOptions({ name: 'InheritPackageFormItem' });
const props = defineProps<{
  bizs?: number | string;
}>();
const emit = defineEmits<{
  (e: 'validateSuccess', host: RollingServerHost): void;
  (e: 'validateFailed'): void;
}>();
const model = defineModel<string>();

const { t } = useI18n();
const { whereAmI, getBizsId } = useWhereAmI();
const { getMonthName } = useCvmChargeType();

const formItem = useTemplateRef('formItem');
const isCheckLoading = ref(false);
const rollingServerHost = ref<RollingServerHost>();

const checkRollingSeverHost = (bk_asset_id: string) => {
  if (!bk_asset_id) return true;
  return new Promise(async (resolve, reject) => {
    isCheckLoading.value = true;
    try {
      const bk_biz_id = Senarios.service === whereAmI.value ? props.bizs : getBizsId();
      const res = await http.post('/api/v1/woa/task/check/rolling_server/host', { bk_biz_id, bk_asset_id });
      rollingServerHost.value = res.data;
      // 将校验成功的机器代表信息回传
      emit('validateSuccess', res.data);
      resolve(res.data);
    } catch (error) {
      // 校验失败
      rollingServerHost.value = null;
      emit('validateFailed');
      reject(error);
    } finally {
      isCheckLoading.value = false;
    }
  });
};
</script>

<template>
  <FormItem
    ref="formItem"
    :label="t('继承套餐的机器代表')"
    property="bk_asset_id"
    required
    :description="t('选择填写业务下一台机器作为继承套餐的代表，继承原套餐类型、计费时长、地域大区等信息。')"
    :rules="[{ validator: (value: string) => checkRollingSeverHost(value), message: t('校验失败')}]"
  >
    <InputWithValidate
      v-model="model"
      class="w600"
      :loading="isCheckLoading"
      @click="formItem.validate()"
      :placeholder="t('请输入一个继承机器(继承套餐时长、计费模式)的CC固资号')"
    />
    <!-- 校验成功，展示继承的机器信息，并且在添加配置清单时设置默认值（计费模式、购买时长...）。 -->
    <ul class="inherit-instance-info" v-if="rollingServerHost">
      <li>
        <span class="label">{{ t('机型：') }}</span>
        <span>{{ rollingServerHost.device_type }}</span>
      </li>
      <li>
        <span class="label">{{ t('计费模式：') }}</span>
        <span>{{ INSTANCE_CHARGE_MAP[rollingServerHost.instance_charge_type] }}</span>
      </li>
      <li>
        <span class="label">{{ t('剩余时间：') }}</span>
        <span>{{ rollingServerHost.charge_months ? getMonthName(rollingServerHost.charge_months) : '--' }}</span>
      </li>
      <li>
        <span class="label">{{ t('计费起始时间：') }}</span>
        <span>{{ timeFormatter(rollingServerHost.billing_start_time) }}</span>
      </li>
      <li>
        <span class="label">{{ t('计费过期时间：') }}</span>
        <span>{{ timeFormatter(rollingServerHost.old_billing_expire_time) }}</span>
      </li>
    </ul>
  </FormItem>
</template>

<style scoped lang="scss">
.w600 {
  width: 600px;
}

.inherit-instance-info {
  display: flex;
  align-items: center;

  li {
    margin-right: 24px;
    font-size: 12px;

    &:last-of-type {
      margin-right: 0;
    }

    .label {
      color: $font-deep-color;
    }
  }
}
</style>
