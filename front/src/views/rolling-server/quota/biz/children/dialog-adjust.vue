<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import dayjs from 'dayjs';
import { useRollingServerQuotaStore, type IRollingServerBizQuotaItem } from '@/store/rolling-server-quota';
import { QuotaAdjustType } from '@/views/rolling-server/typings';
import { quotaAdjustTypeNames } from '@/views/rolling-server/constants';

const props = defineProps<{ dataRow: IRollingServerBizQuotaItem }>();
const model = defineModel<boolean>();

// 当前or跨月，跨月不传入dataRow
const isCurrentMonth = computed(() => props.dataRow?.id !== undefined);
const title = computed(() => (isCurrentMonth.value ? '调整额度' : '跨月调整额度'));

const rollingServerQuotaStore = useRollingServerQuotaStore();

const formRef = ref(null);

const formData = reactive({
  bk_biz_ids: isCurrentMonth.value ? [props.dataRow.bk_biz_id] : [],
  base_quota: isCurrentMonth.value ? props.dataRow.base_quota : undefined,
  adjust_type: isCurrentMonth.value ? props.dataRow.adjust_type : QuotaAdjustType.INCREASE,
  quota_offset: 1,
  adjust_month: isCurrentMonth.value ? [new Date(), new Date()] : [],
});

const disabledDate = (date: Date) => {
  const dateTime = date.valueOf();
  return dateTime < dayjs().add(0, 'month').valueOf() || dateTime > dayjs().add(12, 'month').valueOf();
};

const closeDialog = () => {
  model.value = false;
};

const handleDialogConfirm = async () => {
  await formRef.value?.validate();
  const saveData: any = { ...formData };
  saveData.adjust_month = {
    start: dayjs(saveData.adjust_month[0]).format('YYYY-MM'),
    end: dayjs(saveData.adjust_month[1]).format('YYYY-MM'),
  };
  await rollingServerQuotaStore.adjustBizQuota(saveData);
  closeDialog();
};
</script>

<template>
  <bk-dialog
    :title="title"
    :quick-close="false"
    :is-show="model"
    :is-loading="rollingServerQuotaStore.adjustBizQuotaLoading"
    width="580"
    @confirm="handleDialogConfirm"
    @closed="closeDialog"
  >
    <bk-form form-type="vertical" :model="formData" v-if="isCurrentMonth" ref="formRef">
      <bk-form-item label="业务" property="bk_biz_ids">
        <hcm-form-business v-model="formData.bk_biz_ids" disabled />
      </bk-form-item>
      <bk-form-item label="基础额度" property="bk_biz_ids">
        <hcm-form-number v-model="formData.base_quota" disabled />
      </bk-form-item>
      <bk-form-item label="调整额度" class="compose-form-item">
        <bk-form-item property="adjust_type">
          <hcm-form-enum v-model="formData.adjust_type" :option="quotaAdjustTypeNames" />
        </bk-form-item>
        <bk-form-item property="quota_offset" :required="true" class="flex-item">
          <hcm-form-number
            v-model="formData.quota_offset"
            :min="1"
            :max="rollingServerQuotaStore.globalQuotaConfig.global_quota ?? 100000"
          />
        </bk-form-item>
      </bk-form-item>
      <bk-form-item label="调整后额度">
        <div class="adjust-after">50000</div>
      </bk-form-item>
    </bk-form>
    <bk-form form-type="vertical" :model="formData" ref="formRef" v-else>
      <bk-form-item label="业务" :required="true" property="bk_biz_ids">
        <hcm-form-business v-model="formData.bk_biz_ids" multiple />
      </bk-form-item>
      <bk-form-item label="调整月份" :required="true" property="adjust_month">
        <hcm-form-datetime
          type="monthrange"
          class="flex-width"
          v-model="formData.adjust_month"
          format="yyyy-MM"
          :disabled-date="disabledDate"
        />
      </bk-form-item>
      <bk-form-item label="调整额度" class="compose-form-item">
        <bk-form-item property="adjust_type">
          <hcm-form-enum v-model="formData.adjust_type" :option="quotaAdjustTypeNames" />
        </bk-form-item>
        <bk-form-item property="quota_offset" :required="true" class="flex-item">
          <hcm-form-number
            v-model="formData.quota_offset"
            :min="1"
            :max="rollingServerQuotaStore.globalQuotaConfig.global_quota ?? 100000"
          />
        </bk-form-item>
      </bk-form-item>
      <bk-form-item label="调整后额度">
        <div class="adjust-after">50000</div>
      </bk-form-item>
    </bk-form>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.compose-form-item {
  :deep(> .bk-form-content) {
    display: flex;
    align-items: center;
    gap: 12px;
    .bk-form-item {
      margin-bottom: 0;

      &.flex-item {
        flex: 1;
      }
    }
  }
}
.flex-width {
  width: 100%;
}
.adjust-after {
  height: 32px;
  background: #fdf4e8;
  border-radius: 2px;
  padding: 0 10px;
  font-weight: 700;
  font-size: 14px;
  color: #e38b02;
}
</style>
