<script setup lang="ts">
import { computed, reactive, ref, watchEffect } from 'vue';
import dayjs from 'dayjs';
import { Message } from 'bkui-vue';
import { useRollingServerQuotaStore } from '@/store/rolling-server-quota';
import { type IBusinessItem } from '@/store/business-global';

const emit = defineEmits<{
  'create-success': [res: { id: string }];
}>();

const rollingServerQuotaStore = useRollingServerQuotaStore();

const model = defineModel<boolean>();

const formData = reactive({
  bk_biz_ids: [],
  quota: rollingServerQuotaStore.globalQuotaConfig.biz_quota,
  quota_month: dayjs().format('YYYY-MM'),
});

const quotaMax = computed(() => {
  return rollingServerQuotaStore.globalQuotaConfig.global_quota ?? rollingServerQuotaStore.globalQuotaConfig.biz_quota;
});

const businessOptionDisabled = computed(() => {
  return (option: IBusinessItem) => hasQuotaBizList.value.some((item) => item.bk_biz_id === option.id);
});

const formRef = ref(null);

const hasQuotaBizList = ref([]);

const closeDialog = () => {
  model.value = false;
};

watchEffect(async () => {
  hasQuotaBizList.value = await rollingServerQuotaStore.getExistQuotaBizList({ quota_month: formData.quota_month });
});

const handleDialogConfirm = async () => {
  await formRef.value?.validate();
  const resData = await rollingServerQuotaStore.createBizQuota(formData);
  emit('create-success', resData);
  Message({ theme: 'success', message: '新增成功' });
  closeDialog();
};
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
        <hcm-form-number v-model="formData.quota" :min="1" :max="quotaMax" placeholder="1" />
      </bk-form-item>
    </bk-form>
  </bk-dialog>
</template>
