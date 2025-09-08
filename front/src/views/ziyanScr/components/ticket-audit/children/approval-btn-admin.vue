<script setup lang="ts">
import { computed, inject, onMounted, Ref, ref, useAttrs } from 'vue';
import { useI18n } from 'vue-i18n';
import useFormModel from '@/hooks/useFormModel';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useResSubTicketStore, SubTicketItem, SubTicketDetail } from '@/store/ticket/res-sub-ticket';

interface IProps {
  loading?: boolean;
  confirmHandler: (formModel: IFormModel) => Promise<any>;
}
interface IFormModel {
  approval: boolean;
  use_transfer_pool: boolean;
}

const props = withDefaults(defineProps<IProps>(), {
  loading: false,
});
const emit = defineEmits<(e: 'shown' | 'hidden') => void>();
const attrs = useAttrs();
const store = useResSubTicketStore();
const { t } = useI18n();
const { getBizsId, isBusinessPage } = useWhereAmI();

const ticketDetail = inject<Ref<SubTicketDetail>>('ticketDetail');
const ticketListData = inject<Ref<SubTicketItem>>('ticketListData');

const isShow = ref(false);
const isConfirmLoading = ref(false);
const { formModel, resetForm } = useFormModel<IFormModel>({ approval: true, use_transfer_pool: false });

const handleShown = () => {
  emit('shown');
};
const handleHidden = () => {
  emit('hidden');
};

const handleConfirm = async () => {
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
    getSummaryData();
  }
};

const currentCore = computed(() => {
  return ticketListData.value?.updated_info?.cvm?.cpu_core;
});

// 额度获取和展示
const remainCore = ref(0);
const approvalCore = ref(0);
const getSummaryData = async () => {
  // 获取剩余额度
  const params = {
    obs_project: ticketDetail.value.demands.map((it) => it?.updated_info?.obs_project),
    technical_class: ticketDetail.value.demands.map((it) => it?.updated_info?.cvm?.technical_class),
    year: new Date().getFullYear(),
  };
  const promise = isBusinessPage
    ? store.getTransferQuotaSummaryByBiz(getBizsId(), {
        ...params,
        bk_biz_id: [getBizsId()],
      })
    : store.getTransferQuotaSummary(params);
  const remainRes = await promise;

  // 获取审批额度
  const approvalRes = await store.getTransferQuotaConfigs();

  remainCore.value = remainRes.data?.remain_quota || 0;
  approvalCore.value = approvalRes.data?.audit_quota || 0;
};

onMounted(() => {
  getSummaryData();
});
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
    <div class="info">
      <bk-form label-width="120">
        <bk-form-item :label="t('审批节点：')">管理员审批</bk-form-item>
        <bk-form-item :label="t('当前核数：')">
          <span class="light">{{ currentCore }} 核</span>
        </bk-form-item>
        <bk-form-item :label="t('审批额度：')">
          <span class="light">{{ approvalCore }} 核</span>
        </bk-form-item>
        <bk-form-item :label="t('中转池剩余额度：')">
          <span class="light">{{ remainCore }} 核</span>
        </bk-form-item>
      </bk-form>
    </div>

    <bk-form form-type="vertical" :model="formModel">
      <bk-form-item :label="t('审批意见')" property="approval" required>
        <bk-radio v-model="formModel.approval" :label="true">{{ t('同意') }}</bk-radio>
        <bk-radio v-model="formModel.approval" :label="false">{{ t('拒绝') }}</bk-radio>
      </bk-form-item>
    </bk-form>

    <p class="mt-28">
      <bk-checkbox v-model="formModel.use_transfer_pool">使用中转池额度</bk-checkbox>
    </p>

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

.info {
  width: 100%;
  background-color: #f5f7fa;
  margin-bottom: 20px;
  padding: 8px;

  .light {
    color: #f59500;
  }

  :deep(.bk-form-item) {
    margin-bottom: 8px;
  }

  :deep(.bk-form-label) {
    padding-right: 0;
  }
}
</style>
