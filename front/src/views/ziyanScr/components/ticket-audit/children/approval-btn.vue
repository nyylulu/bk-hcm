<script setup lang="ts">
import { ref, useAttrs, useTemplateRef } from 'vue';
import { useI18n } from 'vue-i18n';
import useFormModel from '@/hooks/useFormModel';

interface IProps {
  loading?: boolean;
  confirmHandler: (formModel: IFormModel) => Promise<any>;
}
interface IFormModel {
  approval: boolean;
  remark: string;
}

const props = withDefaults(defineProps<IProps>(), {
  loading: false,
});
const emit = defineEmits<(e: 'shown' | 'hidden') => void>();
const attrs = useAttrs();
const form = useTemplateRef('form');

const { t } = useI18n();

const isShow = ref(false);
const isConfirmLoading = ref(false);
const { formModel, resetForm } = useFormModel<IFormModel>({ approval: true, remark: '' });

const handleShown = () => {
  emit('shown');
};

const handleHidden = () => {
  emit('hidden');
};

const handleConfirm = async () => {
  await form.value.validate();
  isConfirmLoading.value = true;
  try {
    await props.confirmHandler(formModel);
    isShow.value = false;
    resetForm();
  } catch (error) {
    console.error(error);
    return Promise.reject(error);
  } finally {
    isConfirmLoading.value = false;
  }
};
</script>

<template>
  <bk-button
    class="approval-btn"
    :class="attrs.class"
    size="small"
    theme="primary"
    :loading="loading"
    @click="isShow = true"
  >
    {{ t('立即处理') }}
  </bk-button>
  <bk-dialog v-model:is-show="isShow" :title="t('审批')" @shown="handleShown" @hidden="handleHidden">
    <bk-form ref="form" form-type="vertical" :model="formModel">
      <bk-form-item :label="t('审批意见')" property="approval" required>
        <bk-radio v-model="formModel.approval" :label="true">{{ t('同意') }}</bk-radio>
        <bk-radio v-model="formModel.approval" :label="false">{{ t('拒绝') }}</bk-radio>
      </bk-form-item>
      <bk-form-item :label="t('审批说明')" property="remark" required>
        <bk-input type="textarea" :rows="4" :resize="false" v-model="formModel.remark" />
      </bk-form-item>
    </bk-form>

    <template #footer>
      <bk-button theme="primary" :loading="isConfirmLoading" @click="handleConfirm">
        {{ t('确定') }}
      </bk-button>
      <bk-button @click="isShow = false">
        {{ t('取消') }}
      </bk-button>
    </template>
  </bk-dialog>
</template>

<style scoped lang="scss">
.approval-btn {
  font-size: 14px;
  font-weight: normal;
}

:deep(.bk-dialog-footer) {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;

  .bk-button {
    min-width: 88px;
  }
}
</style>
