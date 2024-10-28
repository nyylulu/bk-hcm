<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import dayjs from 'dayjs';
import { useRollingServerQuotaStore } from '@/store/rolling-server-quota';
import { type IBusinessItem } from '@/store/business-global';

const rollingServerQuotaStore = useRollingServerQuotaStore();

const model = defineModel<boolean>();

const formData = reactive({
  bk_biz_ids: [],
  quota: rollingServerQuotaStore.globalQuotaConfig.biz_quota ?? 1,
  quota_month: dayjs().format('YYYY-MM'),
});

const formRef = ref(null);

const closeDialog = () => {
  model.value = false;
};

const handleDialogConfirm = async () => {
  await formRef.value?.validate();
  await rollingServerQuotaStore.createBizQuota(formData);
  closeDialog();
};

const hasQuotaBizList = ref([{ id: 2005000019, name: 'test' }]);

const businessOptionDisabled = computed(() => {
  return (option: IBusinessItem) => hasQuotaBizList.value.some((item) => item.id === option.id);
});
</script>

<template>
  <bk-dialog
    :title="'新增额度'"
    :quick-close="false"
    :is-show="model"
    :is-loading="rollingServerQuotaStore.createBizQuotaLoading"
    @confirm="handleDialogConfirm"
    @closed="closeDialog"
  >
    <bk-form form-type="vertical" :model="formData" ref="formRef">
      <bk-form-item label="业务" :required="true" property="bk_biz_ids">
        <hcm-form-business v-model="formData.bk_biz_ids" multiple :option-disabled="businessOptionDisabled" />
      </bk-form-item>
      <bk-form-item label="基础额度" :required="true" property="quota">
        <hcm-form-number v-model="formData.quota" :min="1" :max="100000" />
      </bk-form-item>
    </bk-form>
  </bk-dialog>
</template>
